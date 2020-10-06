package hostnode

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

const MinPeersForSyncStart = 3

type peerInfo struct {
	ID              peer.ID
	TipSlot         uint64
	TipHeight       uint64
	TipHash         chainhash.Hash
	JustifiedSlot   uint64
	JustifiedHeight uint64
	JustifiedHash   chainhash.Hash
	FinalizedSlot   uint64
	FinalizedHeight uint64
	FinalizedHash   chainhash.Hash
}

var (
	// ErrorBlockAlreadyKnown returns when received a block already known
	ErrorBlockAlreadyKnown = errors.New("block already known")

	// ErrorBlockParentUnknown returns when received a block with an unknown parent
	ErrorBlockParentUnknown = errors.New("unknown block parent")
)

// synchronizer handles syncing for a blockchain.
type synchronizer struct {
	host HostNode
	ctx  context.Context
	log  logger.Logger

	chain chain.Blockchain

	peersTrack     map[peer.ID]*peerInfo
	peersTrackLock sync.Mutex

	sync     bool
	withPeer peer.ID

	lastFinalizedEpoch uint64
}

// NewSyncronizerl constructs a new sync protocol with a given host and chain.
func NewSyncronizer(host HostNode, chain chain.Blockchain) (*synchronizer, error) {

	sp := &synchronizer{
		host:               host,
		log:                config.GlobalParams.Logger,
		ctx:                config.GlobalParams.Context,
		chain:              chain,
		sync:               true,
		peersTrack:         make(map[peer.ID]*peerInfo),
	}

	if err := host.RegisterHandler(p2p.MsgVersionCmd, sp.handleVersionMsg); err != nil {
		return nil, err
	}

	if err := host.RegisterHandler(p2p.MsgGetBlocksCmd, sp.handleGetBlocksMsg); err != nil {
		return nil, err
	}

	if err := host.RegisterTopicHandler(p2p.MsgBlockCmd, sp.handleBlockMsg); err != nil {
		return nil, err
	}

	if err := host.RegisterHandler(p2p.MsgBlockCmd, sp.handleBlockMsg); err != nil {
		return nil, err
	}

	if err := host.RegisterHandler(p2p.MsgSyncEndCmd, sp.handleSyncEndMsg); err != nil {
		return nil, err
	}

	if err := host.RegisterTopicHandler(p2p.MsgFinalizationCmd, sp.handleFinalizationMsg); err != nil {
		return nil, err
	}

	host.GetHost().Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, conn network.Conn) {
			if conn.Stat().Direction != network.DirOutbound {
				return
			}

			// open a stream for the sync protocol:
			s, err := sp.host.GetHost().NewStream(sp.ctx, conn.RemotePeer(), params.ProtocolID)
			if err != nil {
				sp.log.Errorf("could not open stream for connection: %s", err)
			}

			sp.host.HandleStream(s)

			sp.sendVersion(conn.RemotePeer())
		},
		DisconnectedF: func(n network.Network, conn network.Conn) {
			sp.peersTrackLock.Lock()
			defer sp.peersTrackLock.Unlock()
			delete(sp.peersTrack, conn.RemotePeer())
		},
	})

	go sp.initialBlockDownload()

	return sp, nil
}

func (sp *synchronizer) initialBlockDownload() {

	for {
		time.Sleep(time.Second * 1)
		if len(sp.peersTrack) < MinPeersForSyncStart {
			continue
		}
		break
	}

	sp.peersTrackLock.Lock()
	defer sp.peersTrackLock.Unlock()

	myInfo := sp.versionMsg()

	var peersAhead []*peerInfo
	var peersBehind []*peerInfo
	var peersEqual []*peerInfo

	for _, p := range sp.peersTrack {
		if p.TipHeight > myInfo.FinalizedHeight {
			peersAhead = append(peersAhead, p)
		}

		if p.TipHeight == myInfo.FinalizedHeight {
			peersEqual = append(peersEqual, p)
		}

		if p.TipHeight < myInfo.FinalizedHeight {
			peersBehind = append(peersBehind, p)
		}
	}

	if len(peersAhead) == 0 {
		sp.sync = false
		return
	}

	r := rand.Intn(len(peersAhead))
	peerSelected := peersAhead[r]

	sp.askForBlocks(peerSelected.ID)

}

// askForBlocks will ask a peer for blocks.
func (sp *synchronizer) askForBlocks(id peer.ID) {

	sp.sync = true
	sp.withPeer = id

	finalized, _ := sp.chain.State().GetFinalizedHead()

	err := sp.host.SendMessage(id, &p2p.MsgGetBlocks{
		LastBlockHash: finalized.Hash,
	})

	if err != nil {
		sp.log.Error("unable to send block request msg")
	}
}

func (sp *synchronizer) handleBlockMsg(id peer.ID, msg p2p.Message) error {
	block, ok := msg.(*p2p.MsgBlock)
	if !ok {
		return errors.New("non block msg")
	}

	if sp.sync && sp.withPeer != id {
		return nil
	}
	err := sp.processBlock(block.Data)
	if err != nil {
		if err == ErrorBlockAlreadyKnown {
			sp.log.Error(err)
			return nil
		}
		if err == ErrorBlockParentUnknown {
			if !sp.sync {
				sp.log.Error(err)
				p, ok := sp.peersTrack[id]
				if !ok {
					return nil
				}
				fin, _ := sp.chain.State().GetFinalizedHead()
				if p.FinalizedHeight >= fin.Height {
					sp.askForBlocks(id)
				}
				return nil
			}
			return nil
		}
	}

	return nil
}

func (sp *synchronizer) handleFinalizationMsg(id peer.ID, msg p2p.Message) error {

	fin, ok := msg.(*p2p.MsgFinalization)
	if !ok {
		return errors.New("non block msg")
	}

	if sp.host.GetHost().ID() == id {
		return nil
	}

	sp.peersTrackLock.Lock()
	defer sp.peersTrackLock.Unlock()
	_, ok = sp.peersTrack[id]
	if !ok {
		return nil
	}

	sp.peersTrack[id] = &peerInfo{
		ID:              id,
		TipSlot:         fin.TipSlot,
		TipHeight:       fin.Tip,
		TipHash:         fin.TipHash,
		JustifiedSlot:   fin.JustifiedSlot,
		JustifiedHeight: fin.JustifiedHeight,
		JustifiedHash:   fin.JustifiedHash,
		FinalizedSlot:   fin.FinalizedSlot,
		FinalizedHeight: fin.FinalizedHeight,
		FinalizedHash:   fin.FinalizedHash,
	}

	return nil
}

func (sp *synchronizer) handleSyncEndMsg(id peer.ID, msg p2p.Message) error {
	_, ok := msg.(*p2p.MsgSyncEnd)
	if !ok {
		return errors.New("non syncend msg")
	}
	sp.log.Info("syncing finished")
	if !sp.sync {
		return nil
	}
	if sp.withPeer == id {
		sp.sync = false
		sp.withPeer = ""
	}
	return nil
}

func (sp *synchronizer) handleGetBlocksMsg(id peer.ID, rawMsg p2p.Message) error {
	msg, ok := rawMsg.(*p2p.MsgGetBlocks)
	if !ok {
		return errors.New("did not receive get blocks message")
	}

	sp.log.Debug("received getblocks")

	// Get the announced last block to make sure we have a common point
	firstCommon, ok := sp.chain.State().Index().Get(msg.LastBlockHash)
	if !ok {
		sp.log.Errorf("unable to find common point for peer %s", id)
		return nil
	}

	for {
		ch := sp.chain.State().Chain()
		var ok bool
		firstCommon, ok = ch.Next(firstCommon)
		if !ok {
			break
		}

		block, err := sp.chain.GetBlock(firstCommon.Hash)
		if err != nil {
			return err
		}

		err = sp.host.SendMessage(id, &p2p.MsgBlock{
			Data: block,
		})

		if err != nil {
			return err
		}

	}
	err := sp.host.SendMessage(id, &p2p.MsgSyncEnd{})
	if err != nil {
		return err
	}
	return nil
}

func (sp *synchronizer) handleVersionMsg(id peer.ID, msg p2p.Message) error {
	theirVersion, ok := msg.(*p2p.MsgVersion)
	if !ok {
		return fmt.Errorf("did not receive version message")
	}

	sp.log.Infof("received version message from %s", id)

	// Send our version message if required
	ourVersion := sp.versionMsg()
	direction := sp.host.GetPeerDirection(id)

	if direction == network.DirInbound {
		if err := sp.host.SendMessage(id, ourVersion); err != nil {
			return err
		}

	}

	sp.peersTrackLock.Lock()
	sp.peersTrack[id] = &peerInfo{
		ID:              id,
		TipSlot:         theirVersion.TipSlot,
		TipHeight:       theirVersion.Tip,
		TipHash:         theirVersion.TipHash,
		JustifiedSlot:   theirVersion.JustifiedSlot,
		JustifiedHeight: theirVersion.JustifiedHeight,
		JustifiedHash:   theirVersion.JustifiedHash,
		FinalizedSlot:   theirVersion.FinalizedSlot,
		FinalizedHeight: theirVersion.FinalizedHeight,
		FinalizedHash:   theirVersion.FinalizedHash,
	}
	sp.peersTrackLock.Unlock()

	return nil
}

func (sp *synchronizer) processBlock(block *primitives.Block) error {

	// Check if we already have this block
	if sp.chain.State().Index().Have(block.Hash()) {
		return ErrorBlockAlreadyKnown
	}

	// Check if the parent block is known.
	if !sp.chain.State().Index().Have(block.Header.PrevBlockHash) {
		return ErrorBlockParentUnknown
	}

	// Process block
	sp.log.Debugf("processing block %s", block.Hash())
	if err := sp.chain.ProcessBlock(block); err != nil {
		return err
	}

	// The sync protocol has an internal tracker of the lastFinalizedEpoch to know
	// when a new state is finalized.
	// When this happens we should announce all blocks our new status.

	if sp.chain.State().TipState().GetFinalizedEpoch() > sp.lastFinalizedEpoch && !sp.sync {

		tip := sp.chain.State().Tip()
		justified, _ := sp.chain.State().GetJustifiedHead()
		finalized, _ := sp.chain.State().GetFinalizedHead()

		msg := &p2p.MsgFinalization{
			Tip:             tip.Height,
			TipSlot:         tip.Slot,
			TipHash:         tip.Hash,
			JustifiedSlot:   justified.Slot,
			JustifiedHeight: justified.Height,
			JustifiedHash:   justified.Hash,
			FinalizedSlot:   finalized.Slot,
			FinalizedHeight: finalized.Height,
			FinalizedHash:   finalized.Hash,
		}

		err := sp.host.Broadcast(msg)
		if err != nil {
			sp.log.Error(err)
		}

	}

	sp.lastFinalizedEpoch = sp.chain.State().TipState().GetFinalizedEpoch()

	return nil
}

func (sp *synchronizer) versionMsg() *p2p.MsgVersion {

	justified, _ := sp.chain.State().GetJustifiedHead()
	finalized, _ := sp.chain.State().GetFinalizedHead()

	tip := sp.chain.State().Chain().Tip()

	buf := make([]byte, 8)
	rand.Read(buf)
	msg := &p2p.MsgVersion{
		Tip:             tip.Height,
		TipHash:         tip.Hash,
		Nonce:           binary.LittleEndian.Uint64(buf),
		Timestamp:       uint64(time.Now().Unix()),
		JustifiedSlot:   justified.Slot,
		JustifiedHeight: justified.Height,
		JustifiedHash:   justified.Hash,
		FinalizedSlot:   finalized.Slot,
		FinalizedHeight: finalized.Height,
		FinalizedHash:   finalized.Hash,
	}
	return msg
}

func (sp *synchronizer) sendVersion(id peer.ID) {
	msg := sp.versionMsg()
	err := sp.host.SendMessage(id, msg)
	if err != nil {
		sp.log.Errorf("error sending version message: %s", err)
		_ = sp.host.DisconnectPeer(id)
	}
	return
}
