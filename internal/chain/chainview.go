package chain

import (
	"errors"

	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// ChainView is a view of a certain chain in the block tree so that block processing can access valid blocks.
type ChainView struct {
	tip *chainindex.BlockRow

	// effectiveTipSlot is used when the chain is being updated (excluding blocks)
	effectiveTipSlot uint64
}

// NewChainView creates a new chain view with a certain tip
func NewChainView(tip *chainindex.BlockRow) ChainView {
	return ChainView{tip, tip.Slot}
}

// SetTipSlot sets the effective tip slot (which may be updated due to slot transitions)
func (c *ChainView) SetTipSlot(slot uint64) {
	c.effectiveTipSlot = slot
}

// GetHashBySlot gets a hash of a block in a certain slot.
func (c *ChainView) GetHashBySlot(slot uint64) (chainhash.Hash, error) {
	ancestor := c.tip.GetAncestorAtSlot(slot)
	if ancestor == nil {
		if slot > c.effectiveTipSlot {
			return chainhash.Hash{}, errors.New("could not get block past tip")
		}
		ancestor = c.tip
	}
	return ancestor.Hash, nil
}

// Tip gets the tip of the blockchain.
func (c *ChainView) Tip() (chainhash.Hash, error) {
	return c.tip.Hash, nil
}

// GetLastStateRoot gets the state root of the tip.
func (c *ChainView) GetLastStateRoot() (chainhash.Hash, error) {
	return c.tip.StateRoot, nil
}

var _ primitives.BlockView = (*ChainView)(nil)

// GetSubView gets a view of the blockchain at a certain tip.
func (s *StateService) GetSubView(tip chainhash.Hash) (ChainView, error) {
	tipNode, found := s.blockIndex.Get(tip)
	if !found {
		return ChainView{}, errors.New("could not find tip node")
	}
	return NewChainView(tipNode), nil
}

// Tip gets the tip of the blockchain.
func (s *StateService) Tip() *chainindex.BlockRow {
	return s.blockChain.Tip()
}
