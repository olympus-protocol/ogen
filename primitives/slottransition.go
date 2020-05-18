package primitives

import (
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// ProcessSlot runs a slot transition on state, mutating it.
func (s *State) ProcessSlot(p *params.ChainParams, previousBlockRoot chainhash.Hash) {
	// increase the slot number
	s.Slot++

	s.LatestBlockHashes[(s.Slot-1)%p.LatestBlockRootsLength] = previousBlockRoot
}

// BlockView is the view of the blockchain at a certain tip.
type BlockView interface {
	GetHashBySlot(slot uint64) (chainhash.Hash, error)
	Tip() (chainhash.Hash, error)
	SetTipSlot(slot uint64)
	GetLastStateRoot() (chainhash.Hash, error)
}

// ProcessSlots runs slot and epoch transitions until the state matches the requested
// slot.
func (s *State) ProcessSlots(requestedSlot uint64, view BlockView, p *params.ChainParams, log *logger.Logger) ([]*EpochReceipt, error) {
	totalReceipts := make([]*EpochReceipt, 0)

	for s.Slot < requestedSlot {
		// this only happens when there wasn't a block at the first slot of the epoch
		if s.Slot/p.EpochLength > s.EpochIndex && s.Slot%p.EpochLength == 0 {
			receipts, err := s.ProcessEpochTransition(p, log)
			if err != nil {
				return nil, err
			}
			totalReceipts = append(totalReceipts, receipts...)
		}

		tip, err := view.Tip()
		if err != nil {
			return nil, err
		}

		s.ProcessSlot(p, tip)

		view.SetTipSlot(s.Slot)
	}

	return totalReceipts, nil
}
