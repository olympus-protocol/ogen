package state

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// SetSlot runs a slot transition on state, mutating it.
func (s *state) SetSlot(slot uint64) {
	s.Slot = slot
}

// ProcessSlot runs a slot transition on state, mutating it.
func (s *state) ProcessSlot(previousBlockRoot chainhash.Hash) {
	netParams := config.GlobalParams.NetParams

	// increase the slot number
	s.Slot++

	s.LatestBlockHashes[(s.Slot-1)%netParams.LatestBlockRootsLength] = previousBlockRoot
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
func (s *state) ProcessSlots(requestedSlot uint64, view BlockView) ([]*primitives.EpochReceipt, error) {
	netParams := config.GlobalParams.NetParams

	totalReceipts := make([]*primitives.EpochReceipt, 0)

	for s.Slot < requestedSlot {
		// this only happens when there wasn't a block at the first slot of the epoch
		if s.Slot/netParams.EpochLength > s.EpochIndex && s.Slot%netParams.EpochLength == 0 {
			receipts, err := s.ProcessEpochTransition()
			if err != nil {
				return nil, err
			}
			totalReceipts = append(totalReceipts, receipts...)
		}

		tip, err := view.Tip()
		if err != nil {
			return nil, err
		}

		s.ProcessSlot(tip)

		view.SetTipSlot(s.Slot)
	}

	return totalReceipts, nil
}
