package host

import (
	"context"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"time"
)

const MinPeersForSyncStart = 3

var (
	// ErrorBlockAlreadyKnown returns when received a block already known
	ErrorBlockAlreadyKnown = errors.New("block already known")

	// ErrorBlockParentUnknown returns when received a block with an unknown parent
	ErrorBlockParentUnknown = errors.New("unknown block parent")
)

type synchronizer struct {
	host Host
	ctx  context.Context
	log  logger.Logger

	chain chain.Blockchain

	synced          bool
	withPeer        peer.ID
	blockStallTimer *time.Timer

	lastFinalizedEpoch uint64
}

func (sp *synchronizer) initialBlockDownload() {

	for {
		time.Sleep(time.Second * 1)
		if sp.host.TrackedPeers() < MinPeersForSyncStart {
			continue
		}
		break
	}

	peerSelected, ok := sp.host.FindBestPeer()
	if !ok {
		sp.synced = true
		return
	}

	sp.askForBlocks(peerSelected)

	return
}

// askForBlocks will ask a peer for blocks.
func (sp *synchronizer) askForBlocks(id peer.ID) {

	sp.synced = false
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
	sp.synced = true
	sp.withPeer = ""
	sp.log.Info("sync finished")
	return
}

func (sp *synchronizer) handleVersionMsg(id peer.ID, msg p2p.Message) error {
	theirVersion, ok := msg.(*p2p.MsgVersion)
	if !ok {
		return fmt.Errorf("did not receive version message")
	}

	sp.log.Infof("received version message from %s", id)

	// Send our version message if required
	ourVersion := sp.host.Version()
	direction := sp.host.GetPeerDirection(id)

	sp.host.AddPeerStats(id, theirVersion, direction)

	if direction == network.DirInbound {
		if err := sp.host.SendMessage(id, ourVersion); err != nil {
			return err
		}

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
		err := fmt.Sprintf("unable to find common point for peer %s", id)
		sp.log.Error(err)
		return nil
	}

	blockRow, ok := sp.chain.State().Chain().Next(firstCommon)
	if !ok {
		err := fmt.Sprintf("unable to next block from common point for peer %s", id)
		sp.log.Error(err)
		return nil
	}

	for {

		block, err := sp.chain.GetBlock(blockRow.Hash)
		if err != nil {
			return nil
		}

		err = sp.host.SendMessage(id, &p2p.MsgBlock{
			Data: block,
		})

		if err != nil {
			return nil
		}

		blockRow, ok = sp.chain.State().Chain().Next(blockRow)
		if !ok {
			break
		}

	}

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

	if sp.chain.State().TipState().GetFinalizedEpoch() > sp.lastFinalizedEpoch && sp.synced {

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

func (sp *synchronizer) handleBlockMsg(id peer.ID, msg p2p.Message) error {
	block, ok := msg.(*p2p.MsgBlock)
	if !ok {
		return errors.New("non block msg")
	}
	if !sp.synced && sp.withPeer != id {
		sp.log.Info("received block during sync, waiting to finish...")
		return nil
	}
	err := sp.processBlock(block.Data)
	if err != nil {
		if err == ErrorBlockAlreadyKnown {
			sp.log.Error(err)
			return nil
		}
		if err == ErrorBlockParentUnknown {
			if sp.synced {
				sp.log.Error(err)
				s, ok := sp.host.GetPeerStats(id)
				if !ok {
					return nil
				}
				just, _ := sp.chain.State().GetJustifiedHead()
				if s.ChainStats.JustifiedSlot >= just.Slot {
					go sp.initialBlockDownload()
					return nil
				}
				return nil
			}
			return nil
		}
		sp.log.Error(err)
		return err
	}

	if !sp.synced {
		sp.blockStallTimer.Reset(time.Second * 3)
	}

	return nil
}

// NewSynchronizer constructs a new sync protocol with a given host and chain.
func NewSynchronizer(host Host, chain chain.Blockchain) (*synchronizer, error) {

	sp := &synchronizer{
		host:   host,
		log:    config.GlobalParams.Logger,
		ctx:    config.GlobalParams.Context,
		chain:  chain,
		synced: false,
	}

	host.RegisterHandler(p2p.MsgVersionCmd, sp.handleVersionMsg)
	host.RegisterHandler(p2p.MsgGetBlocksCmd, sp.handleGetBlocksMsg)
	host.RegisterTopicHandler(p2p.MsgBlockCmd, sp.handleBlockMsg)
	host.RegisterHandler(p2p.MsgBlockCmd, sp.handleBlockMsg)

	go sp.initialBlockDownload()

	return sp, nil
}
