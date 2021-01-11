package hostnode

import (
	"context"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

const MinPeersForSyncStart = 3

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

	sync            bool
	withPeer        peer.ID
	blockStallTimer *time.Timer

	lastFinalizedEpoch uint64
}

// NewSyncronizer constructs a new sync protocol with a given host and chain.
func NewSyncronizer(host HostNode, chain chain.Blockchain) (*synchronizer, error) {

	sp := &synchronizer{
		host:  host,
		log:   config.GlobalParams.Logger,
		ctx:   config.GlobalParams.Context,
		chain: chain,
		sync:  true,
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

	host.GetHost().Network().Notify(&network.NotifyBundle{
		DisconnectedF: func(n network.Network, conn network.Conn) {
			sp.host.StatsService().Remove(conn.RemotePeer())
			n.Close()
			conn.Close()
		},
	})

	go sp.initialBlockDownload()

	return sp, nil
}

func (sp *synchronizer) initialBlockDownload() {

	for {
		time.Sleep(time.Second * 1)
		if sp.host.StatsService().Count() < MinPeersForSyncStart {
			continue
		}
		break
	}

	peerSelected, ok := sp.host.StatsService().FindBestPeer()
	if !ok {
		sp.sync = false
		return
	}

	sp.askForBlocks(peerSelected)

	return
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
		return
	}

	sp.blockStallTimer = time.NewTimer(time.Second * 5)

	go sp.waitForBlocksTimer()

	return
}

func (sp *synchronizer) waitForBlocksTimer() {
	<-sp.blockStallTimer.C
	sp.sync = false
	sp.withPeer = ""
	sp.log.Info("sync finished")
	return
}

func (sp *synchronizer) handleBlockMsg(id peer.ID, msg p2p.Message) (uint64, error) {
	block, ok := msg.(*p2p.MsgBlock)
	if !ok {
		return msg.PayloadLength(), errors.New("non block msg")
	}
	if sp.sync && sp.withPeer != id {
		sp.log.Info("received block during sync, waiting to finish...")
		return msg.PayloadLength(), nil
	}
	err := sp.processBlock(block.Data)
	if err != nil {
		if err == ErrorBlockAlreadyKnown {
			sp.log.Error(err)
			return msg.PayloadLength(), nil
		}
		if err == ErrorBlockParentUnknown {
			if !sp.sync {
				sp.log.Error(err)
				stats, ok := sp.host.StatsService().GetPeerStats(id)
				if !ok {
					return msg.PayloadLength(), nil
				}
				just, _ := sp.chain.State().GetJustifiedHead()
				if stats.ChainStats.JustifiedSlot >= just.Slot {
					go sp.initialBlockDownload()
					return msg.PayloadLength(), nil
				}
				return msg.PayloadLength(), nil
			}
			return msg.PayloadLength(), nil
		}
		sp.log.Error(err)
		return msg.PayloadLength(), err
	}

	if sp.sync {
		sp.blockStallTimer.Reset(time.Second * 3)
	}

	return msg.PayloadLength(), nil
}

func (sp *synchronizer) handleGetBlocksMsg(id peer.ID, rawMsg p2p.Message) (uint64, error) {
	msg, ok := rawMsg.(*p2p.MsgGetBlocks)
	if !ok {
		return 0, errors.New("did not receive get blocks message")
	}

	sp.log.Debug("received getblocks")

	// Get the announced last block to make sure we have a common point
	firstCommon, ok := sp.chain.State().Index().Get(msg.LastBlockHash)
	if !ok {
		err := fmt.Sprintf("unable to find common point for peer %s", id)
		sp.log.Error(err)
		return msg.PayloadLength(), nil
	}

	blockRow, ok := sp.chain.State().Chain().Next(firstCommon)
	if !ok {
		err := fmt.Sprintf("unable to next block from common point for peer %s", id)
		sp.log.Error(err)
		return msg.PayloadLength(), nil
	}

	for {

		block, err := sp.chain.GetBlock(blockRow.Hash)
		if err != nil {
			return msg.PayloadLength(), nil
		}

		err = sp.host.SendMessage(id, &p2p.MsgBlock{
			Data: block,
		})

		if err != nil {
			return msg.PayloadLength(), nil
		}

		blockRow, ok = sp.chain.State().Chain().Next(blockRow)
		if !ok {
			break
		}

	}

	return msg.PayloadLength(), nil
}

func (sp *synchronizer) handleVersionMsg(id peer.ID, msg p2p.Message) (uint64, error) {
	theirVersion, ok := msg.(*p2p.MsgVersion)
	if !ok {
		return 0, fmt.Errorf("did not receive version message")
	}

	sp.log.Infof("received version message from %s", id)

	// Send our version message if required
	ourVersion := sp.host.VersionMsg()
	direction := sp.host.GetPeerDirection(id)

	sp.host.StatsService().Add(id, theirVersion, direction)

	if direction == network.DirInbound {
		if err := sp.host.SendMessage(id, ourVersion); err != nil {
			return msg.PayloadLength(), err
		}

	}

	return msg.PayloadLength(), nil
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

func (sp *synchronizer) sendVersion(id peer.ID) {
	msg := sp.host.VersionMsg()
	err := sp.host.SendMessage(id, msg)
	if err != nil {
		sp.log.Errorf("error sending version message: %s", err)
		_ = sp.host.DisconnectPeer(id)
	}
	return
}
