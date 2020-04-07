package chain

import (
	"sync"

	"github.com/olympus-protocol/ogen/chain/index"
)

type Chain struct {
	lock  sync.RWMutex
	chain []*index.BlockRow
}

func (c *Chain) Height() int32 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return int32(len(c.chain) - 1)
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
	if int32(cap(c.chain)) < needed {
		newChain := make([]*index.BlockRow, needed, 1000+needed)
		copy(newChain, c.chain)
		c.chain = newChain
	} else {
		prevLen := int32(len(c.chain))
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

func (c *Chain) GetNodeByHeight(height int32) (*index.BlockRow, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if height >= int32(len(c.chain)) {
		return nil, false
	}

	return c.chain[height], true
}

// NewBlockchain creates a new chain.
func NewChain(genesisBlock *index.BlockRow) *Chain {
	chain := make([]*index.BlockRow, 1, 1000)
	chain[0] = genesisBlock
	return &Chain{
		chain: chain,
	}
}
