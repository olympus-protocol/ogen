package chain

import (
	"fmt"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"sync"
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

// ChainView keeps track of the current state of the blockchain.
type ChainView struct {
	blockIndex *index.BlockIndex
	blockChain *Chain
}

// NewChainView creates a new chain view.
func NewChainView(genesisHeader p2p.BlockHeader, genesisLocator blockdb.BlockLocation) (*ChainView, error) {
	blockIndex, err := index.InitBlocksIndex(genesisHeader, genesisLocator)
	if err != nil {
		return nil, err
	}

	genesisHash := genesisHeader.Hash()
	row, _ := blockIndex.Get(genesisHash)

	return &ChainView{
		blockIndex: blockIndex,
		blockChain: NewChain(row),
	}, nil
}

func (c *ChainView) Height() int32 {
	return c.blockChain.Height()
}

func (c *ChainView) Add(header p2p.BlockHeader, locator blockdb.BlockLocation) (*index.BlockRow, error) {
	return c.blockIndex.Add(header, locator)
}

func (c *ChainView) Tip() *index.BlockRow {
	return c.blockChain.Tip()
}

func (c *ChainView) SetTip(h chainhash.Hash) error {
	if row, found := c.blockIndex.Get(h); found {
		c.blockChain.SetTip(row)
	}
	return fmt.Errorf("error setting block tip: could not find block with hash: %s", h)
}

func (c *ChainView) GetRowByHeight(height int32) (*index.BlockRow, bool) {
	return c.blockChain.GetNodeByHeight(height)
}

func (c *ChainView) GetRowByHash(h chainhash.Hash) (*index.BlockRow, bool) {
	return c.blockIndex.Get(h)
}

func (c *ChainView) Has(h chainhash.Hash) bool {
	return c.blockIndex.Have(h)
}

var _ ChainInterface = &ChainView{}

// ChainInterface is an interface that allows basic access to the block index and chain.
type ChainInterface interface {
	Add(header p2p.BlockHeader, locator blockdb.BlockLocation) (*index.BlockRow, error)
	Tip() *index.BlockRow
	SetTip(chainhash.Hash) error
	GetRowByHeight(int32) (*index.BlockRow, bool)
	GetRowByHash(chainhash.Hash) (*index.BlockRow, bool)
	Height() int32
	Has(hash chainhash.Hash) bool
}
