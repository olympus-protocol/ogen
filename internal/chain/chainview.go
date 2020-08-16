package chain

import (
	"errors"
	"github.com/olympus-protocol/ogen/internal/state"

	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// View is a view of a certain chain in the block tree so that block processing can access valid blocks.
type View struct {
	tip *chainindex.BlockRow

	// effectiveTipSlot is used when the chain is being updated (excluding blocks)
	effectiveTipSlot uint64
}

// NewChainView creates a new chain view with a certain tip
func NewChainView(tip *chainindex.BlockRow) View {
	return View{tip, tip.Slot}
}

// SetTipSlot sets the effective tip slot (which may be updated due to slot transitions)
func (c *View) SetTipSlot(slot uint64) {
	c.effectiveTipSlot = slot
}

// GetHashBySlot gets a hash of a block in a certain slot.
func (c *View) GetHashBySlot(slot uint64) (chainhash.Hash, error) {
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
func (c *View) Tip() (chainhash.Hash, error) {
	return c.tip.Hash, nil
}

// GetLastStateRoot gets the state root of the tip.
func (c *View) GetLastStateRoot() (chainhash.Hash, error) {
	return c.tip.StateRoot, nil
}

var _ state.BlockView = (*View)(nil)
