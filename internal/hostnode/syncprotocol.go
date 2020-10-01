package hostnode

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
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

// SyncProtocol handles syncing for a blockchain.
type syncProtocol struct {
	host HostNode
	ctx  context.Context
	log  logger.Logger

	chain chain.Blockchain

	protocolHandler *protocolHandler

	peersTrack     map[peer.ID]*peerInfo
	peersTrackLock sync.Mutex

	sync     bool
	withPeer peer.ID

	lastFinalizedEpoch uint64
	unknownBlocksCount uint64
}

func listenToTopic(ctx context.Context, subscription *pubsub.Subscription, handler func(data []byte, id peer.ID)) {
	for {
		msg, err := subscription.Next(ctx)
		if err != nil {
			break
		}

		handler(msg.Data, msg.GetFrom())
	}
}

// NewSyncProtocol constructs a new sync protocol with a given host and chain.
func NewSyncProtocol(host HostNode, chain chain.Blockchain) (*syncProtocol, error) {

	ph, err := newProtocolHandler(params.SyncProtocolID, host)
	if err != nil {
		return nil, err
	}
	sp := &syncProtocol{
		host:               host,
		log:                config.GlobalParams.Logger,
		ctx:                config.GlobalParams.Context,
		protocolHandler:    ph,
		chain:              chain,
		sync:               true,
		peersTrack:         make(map[peer.ID]*peerInfo),
		unknownBlocksCount: 0,
	}

	if err := ph.RegisterHandler(p2p.MsgVersionCmd, sp.handleVersion); err != nil {
		return nil, err
	}

	if err := ph.RegisterHandler(p2p.MsgGetBlocksCmd, sp.handleGetBlocks); err != nil {
		return nil, err
	}

	if err := ph.RegisterHandler(p2p.MsgBlockCmd, sp.blockHandler); err != nil {
		return nil, err
	}

	if err := ph.RegisterHandler(p2p.MsgSyncEndCmd, sp.syncEndHandler); err != nil {
		return nil, err
	}

	if err := ph.RegisterHandler(p2p.MsgFinalizationCmd, sp.handleFinalization); err != nil {
		return nil, err
	}

	if err := sp.listenForFinalizations(); err != nil {
		return nil, err
	}

	if err := sp.listenForBroadcasts(); err != nil {
		return nil, err
	}

	go sp.initialBlockDownload()

	return sp, nil
}

func (sp *syncProtocol) initialBlockDownload() {

	sp.peersTrackLock.Lock()
	defer sp.peersTrackLock.Unlock()

	for {
		time.Sleep(time.Second * 1)
		if len(sp.peersTrack) < MinPeersForSyncStart {
			fmt.Println(len(sp.peersTrack))
			continue
		}
		break
	}

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
func (sp *syncProtocol) askForBlocks(id peer.ID) {

	sp.sync = true
	sp.withPeer = id

	finalized, _ := sp.chain.State().GetFinalizedHead()

	err := sp.protocolHandler.SendMessage(id, &p2p.MsgGetBlocks{
		LastBlockHash: finalized.Hash,
	})

	if err != nil {
		sp.log.Error("unable to send block request msg")
	}
}

func (sp *syncProtocol) listenForBroadcasts() error {
	blockTopic, err := sp.host.Topic(p2p.MsgBlockCmd)
	if err != nil {
		return err
	}

	blockSub, err := blockTopic.Subscribe()
	if err != nil {
		return err
	}

	go listenToTopic(sp.ctx, blockSub, func(data []byte, id peer.ID) {
		if id == sp.host.GetHost().ID() {
			return
		}

		buf := bytes.NewBuffer(data)

		msg, err := p2p.ReadMessage(buf, sp.host.GetNetMagic())

		if err != nil {
			sp.log.Errorf("error decoding msg from peer %s: %s", id, err)
			return
		}

		block, ok := msg.(*p2p.MsgBlock)
		if !ok {
			sp.log.Errorf("wrong message type on block subscription from peer %s: %s", id, err)
			return
		}

		if err := sp.handleBlock(id, block.Data); err != nil {
			sp.log.Errorf("error handling incoming block from peer: %s", err)
		}
	})

	return nil
}

func (sp *syncProtocol) listenForFinalizations() error {
	finTopic, err := sp.host.Topic(p2p.MsgFinalizationCmd)
	if err != nil {
		return err
	}

	finSub, err := finTopic.Subscribe()
	if err != nil {
		return err
	}

	go listenToTopic(sp.ctx, finSub, func(data []byte, id peer.ID) {
		if id == sp.host.GetHost().ID() {
			return
		}

		buf := bytes.NewBuffer(data)

		msg, err := p2p.ReadMessage(buf, sp.host.GetNetMagic())

		if err != nil {
			sp.log.Errorf("error decoding msg from peer %s", id)
			return
		}

		if err := sp.handleFinalization(id, msg); err != nil {
			sp.log.Errorf("error handling incoming finalization from peer: %s", err)
		}
	})
	return nil
}

func (sp *syncProtocol) blockHandler(id peer.ID, msg p2p.Message) error {
	bmsg, ok := msg.(*p2p.MsgBlock)
	if !ok {
		return errors.New("non block msg")
	}
	return sp.handleBlock(id, bmsg.Data)
}

func (sp *syncProtocol) syncEndHandler(id peer.ID, msg p2p.Message) error {
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

func (sp *syncProtocol) handleBlock(id peer.ID, block *primitives.Block) error {
	if sp.sync && sp.withPeer != id {
		return nil
	}
	err := sp.processBlock(block)
	if err != nil {
		if err == ErrorBlockAlreadyKnown {
			sp.log.Error(err)
			return nil
		}
		if err == ErrorBlockParentUnknown {
			if !sp.sync {
				sp.log.Error(err)
				go sp.initialBlockDownload()
				return nil
			}
			return nil
		}
	}
	return nil
}

func (sp *syncProtocol) handleFinalization(id peer.ID, msg p2p.Message) error {

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

func (sp *syncProtocol) processBlock(block *primitives.Block) error {

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

	if sp.chain.State().TipState().GetFinalizedEpoch() > sp.lastFinalizedEpoch && !sp.Syncing() {

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

		err := sp.protocolHandler.SendFinalizedMessage(msg)
		if err != nil {
			sp.log.Error(err)
		}
	}

	sp.lastFinalizedEpoch = sp.chain.State().TipState().GetFinalizedEpoch()

	return nil
}

func (sp *syncProtocol) handleGetBlocks(id peer.ID, rawMsg p2p.Message) error {
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
		var ok bool
		firstCommon, ok = sp.chain.State().Chain().Next(firstCommon)
		if !ok {
			break
		}

		block, err := sp.chain.GetBlock(firstCommon.Hash)
		if err != nil {
			return err
		}

		err = sp.protocolHandler.SendMessage(id, &p2p.MsgBlock{
			Data: block,
		})

		if err != nil {
			return err
		}

	}
	err := sp.protocolHandler.SendMessage(id, &p2p.MsgSyncEnd{})
	if err != nil {
		return err
	}
	return nil
}

func (sp *syncProtocol) handleVersion(id peer.ID, msg p2p.Message) error {
	theirVersion, ok := msg.(*p2p.MsgVersion)
	if !ok {
		return fmt.Errorf("did not receive version message")
	}

	sp.log.Infof("received version message from %s", id)

	// Send our version message if required
	ourVersion := sp.versionMsg()
	//direction := sp.host.GetPeerDirection(id)
	if err := sp.protocolHandler.SendMessage(id, ourVersion); err != nil {
		return err
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

func (sp *syncProtocol) versionMsg() *p2p.MsgVersion {

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

func (sp *syncProtocol) sendVersion(id peer.ID) {
	msg := sp.versionMsg()
	err := sp.protocolHandler.SendMessage(id, msg)
	if err != nil {
		sp.log.Errorf("error sending version message: %s", err)
		_ = sp.host.DisconnectPeer(id)
	}
	return
}

func (sp *syncProtocol) Syncing() bool {
	return sp.sync
}
