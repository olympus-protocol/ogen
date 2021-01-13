package host

import (
	"context"
	"errors"
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
	recentSynced    bool
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
	sp.recentSynced = true
	sp.withPeer = ""
	sp.log.Info("Sync finished. Waiting for the next block...")
	return
}

func (sp *synchronizer) handleBlock(id peer.ID, block *primitives.Block) error {
	if !sp.synced && sp.withPeer != id {
		sp.log.Info("received block during sync, waiting to finish...")
		return nil
	}
	err := sp.processBlock(block)
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

	if sp.recentSynced {
		sp.recentSynced = false
	}

	if !sp.synced {
		sp.blockStallTimer.Reset(time.Second * 3)
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

// NewSynchronizer constructs a new sync protocol with a given host and chain.
func NewSynchronizer(host Host, chain chain.Blockchain) (*synchronizer, error) {

	sp := &synchronizer{
		host:   host,
		log:    config.GlobalParams.Logger,
		ctx:    config.GlobalParams.Context,
		chain:  chain,
		synced: false,
	}

	go sp.initialBlockDownload()

	return sp, nil
}
