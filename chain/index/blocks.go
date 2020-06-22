package index

import (
	"fmt"
	"sync"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// BlockRow represents a single row in the block index.
type BlockRow struct {
	StateRoot chainhash.Hash
	Height    uint64
	Slot      uint64
	Hash      chainhash.Hash
	Parent    *BlockRow

	children     []*BlockRow
	childrenLock sync.RWMutex
}

// AddChild adds a child to the block row.
func (br *BlockRow) AddChild(child *BlockRow) {
	br.childrenLock.Lock()
	defer br.childrenLock.Unlock()

	for _, c := range br.children {
		if c.Hash.IsEqual(&child.Hash) {
			return
		}
	}

	br.children = append(br.children, child)
}

// Children gets the children of the block row.
func (br *BlockRow) Children() []*BlockRow {
	childrenCopy := make([]*BlockRow, len(br.children))
	br.childrenLock.RLock()
	defer br.childrenLock.RUnlock()
	copy(childrenCopy, br.children)
	return childrenCopy
}

// GetAncestorAtSlot gets the block row ancestor at a certain slot.
func (br *BlockRow) GetAncestorAtSlot(slot uint64) *BlockRow {
	if br.Slot < slot {
		return nil
	}

	current := br

	// go up to the slot after the slot we're searching for
	for slot < current.Slot {
		current = current.Parent
	}
	return current
}

// GetAncestorAtHeight gets the block row ancestor at a certain height.
func (br *BlockRow) GetAncestorAtHeight(height uint64) *BlockRow {
	if br.Height < height {
		return nil
	}

	current := br

	// go up to the slot after the slot we're searching for
	for height < current.Height {
		current = current.Parent
	}
	return current
}

var zeroHash = chainhash.Hash{}

// BlockIndex is an index from hash to BlockRow.
type BlockIndex struct {
	lock  sync.RWMutex
	index map[chainhash.Hash]*BlockRow
}

// LoadBlockNode loads a block node and connects it to the parent block.
func (i *BlockIndex) LoadBlockNode(row *bdb.BlockNodeDisk) (*BlockRow, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	parent, found := i.get(row.Parent)
	if !found && row.Slot != 0 {
		return nil, fmt.Errorf("missing parent block %s", row.Parent)
	}

	newNode := &BlockRow{
		Hash:      row.Hash,
		Height:    row.Height,
		Slot:      row.Slot,
		StateRoot: row.StateRoot,
		Parent:    parent,
		children:  make([]*BlockRow, 0),
	}

	i.index[row.Hash] = newNode

	if parent != nil {
		parent.children = append(parent.children, newNode)
	}

	return newNode, nil
}

func (i *BlockIndex) get(hash chainhash.Hash) (*BlockRow, bool) {
	row, found := i.index[hash]
	return row, found
}

// Get gets a block from the block index.
func (i *BlockIndex) Get(hash chainhash.Hash) (*BlockRow, bool) {
	i.lock.RLock()
	defer i.lock.RUnlock()

	row, found := i.get(hash)
	return row, found
}

// Have checks if the block index contains a certain hash.
func (i *BlockIndex) Have(hash chainhash.Hash) bool {
	i.lock.RLock()
	_, ok := i.index[hash]
	i.lock.RUnlock()
	return ok
}

func (i *BlockIndex) add(row *BlockRow) error {
	i.index[row.Hash] = row
	return nil
}

// Add adds a row to the block index.
func (i *BlockIndex) Add(block primitives.Block) (*BlockRow, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	prev, found := i.index[block.Header.PrevBlockHash]
	if !found {
		return nil, fmt.Errorf("could not add block to index: could not find parent with hash %s", block.Header.PrevBlockHash)
	}
	hash, err := block.Header.Hash()
	if err != nil {
		return nil, err
	}
	row := &BlockRow{
		StateRoot: block.Header.StateRoot,
		Height:    prev.Height + 1,
		Parent:    prev,
		Hash:      hash,
		Slot:      block.Header.Slot,
		children:  make([]*BlockRow, 0),
	}

	prev.AddChild(row)

	i.index[block.Header.PrevBlockHash] = prev

	err = i.add(row)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// InitBlocksIndex creates a new block index.
func InitBlocksIndex(genesisBlock primitives.Block) (*BlockIndex, error) {
	headerHash, err := genesisBlock.Header.Hash()
	if err != nil {
		return nil, err
	}
	return &BlockIndex{
		index: map[chainhash.Hash]*BlockRow{
			headerHash: {
				Height: 0,
				Parent: nil,
				Hash:   headerHash,
			},
		},
	}, nil
}

// ToBlockNodeDisk converts an in-memory representation of a block row
// to a serializable version.
func (br *BlockRow) ToBlockNodeDisk() *bdb.BlockNodeDisk {
	childrenNodes := br.Children()
	children := make([]chainhash.Hash, len(childrenNodes))
	for i := range childrenNodes {
		children[i] = childrenNodes[i].Hash
	}

	parent := chainhash.Hash{}
	if br.Parent != nil {
		parent = br.Parent.Hash
	}

	return &bdb.BlockNodeDisk{
		StateRoot: br.StateRoot,
		Height:    br.Height,
		Slot:      br.Slot,
		Children:  children,
		Hash:      br.Hash,
		Parent:    parent,
	}
}
