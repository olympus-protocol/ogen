package chain

import (
	"sync"

	"github.com/olympus-protocol/ogen/chain/index"
)

type Chain struct {
	lock  sync.RWMutex
	chain []*index.BlockRow
}

func (c *Chain) Height() uint64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return uint64(len(c.chain) - 1)
}

// SetTip sets the tip of the chain.
func (c *Chain) SetTip(row *index.BlockRow) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if row == nil {
		c.chain = make([]*index.BlockRow, 0)
		return
	}

	needed := row.Height + 1

	// algorithm copied from btcd chainview
	if uint64(cap(c.chain)) < needed {
		newChain := make([]*index.BlockRow, needed, 1000+needed)
		copy(newChain, c.chain)
		c.chain = newChain
	} else {
		prevLen := uint64(len(c.chain))
		c.chain = c.chain[0:needed]
		for i := prevLen; i < needed; i++ {
			c.chain[i] = nil
		}
	}

	for row != nil && c.chain[row.Height] != row {
		c.chain[row.Height] = row
		row = row.Parent
	}
}

func (c *Chain) Tip() *index.BlockRow {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.chain[len(c.chain)-1]
}

func (c *Chain) Genesis() *index.BlockRow {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.chain[0]
}

func (c *Chain) Next(row *index.BlockRow) (*index.BlockRow, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if uint64(len(c.chain)) <= row.Height+1 {
		return nil, false
	}

	return c.chain[row.Height+1], true
}

func (c *Chain) GetNodeByHeight(height uint64) (*index.BlockRow, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if height >= uint64(len(c.chain)) {
		return nil, false
	}

	return c.chain[height], true
}

// GetNodeBySlot returns the node at a specific slot.
func (c *Chain) GetNodeBySlot(slot uint64) (*index.BlockRow, bool) {
	tip := c.Tip()
	if tip == nil {
		return nil, false
	}
	if tip.Slot < slot {
		return tip, true
	}
	return tip.GetAncestorAtSlot(slot), true
}

// NewBlockchain creates a new chain.
func NewChain(genesisBlock *index.BlockRow) *Chain {
	chain := make([]*index.BlockRow, 1, 1000)
	chain[0] = genesisBlock
	return &Chain{
		chain: chain,
	}
}
