package primitives

import (
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// ProcessSlot runs a slot transition on state, mutating it.
func (s *State) ProcessSlot(p *params.ChainParams, previousBlockRoot chainhash.Hash) {
	// increase the slot number
	s.Slot++

	s.LatestBlockHashes[(s.Slot-1)%p.LatestBlockRootsLength] = previousBlockRoot
}

// ProcessSlots runs epoch and slot transitions until a desired slot.
func (s *State) ProcessSlots(p *params.ChainParams, desiredSlot uint64, lastBlockHash chainhash.Hash) {
	for s.Slot < desiredSlot {
		// if we haven't processed enough epochs and this is the first slot of the epoch,
		// do an epoch transition
		if s.Slot/p.EpochLength > s.EpochIndex && s.Slot%p.EpochLength == 0 {
			// TODO: epoch transition
		}

		s.ProcessSlot(p, lastBlockHash)
	}
}
