package index

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/olympus-protocol/ogen/primitives"

	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// BlockRow represents a single row in the block index.
type BlockRow struct {
	Header     primitives.BlockHeader
	Locator    blockdb.BlockLocation
	Height     int32
	parentHash chainhash.Hash
	Hash       chainhash.Hash
	Parent     *BlockRow
}

// Serialize serializes a block index row to the writer.
func (br *BlockRow) Serialize(w io.Writer) error {
	err := br.Locator.Serialize(w)
	if err != nil {
		return err
	}
	err = br.Header.Serialize(w)
	if err != nil {
		return err
	}

	parentHash := chainhash.Hash{}
	if br.Parent != nil {
		parentHash = br.Parent.Header.Hash()
	}
	err = serializer.WriteElements(w, br.Height, parentHash)
	if err != nil {
		return err
	}
	return nil
}

// Deserialize deserializes a block row from the provided reader.
func (br *BlockRow) Deserialize(r io.Reader) error {
	err := br.Locator.Deserialize(r)
	if err != nil {
		return err
	}
	err = br.Header.Deserialize(r)
	if err != nil {
		return err
	}
	err = serializer.ReadElements(r, &br.Height, &br.parentHash)
	if err != nil {
		return err
	}
	br.Hash = br.Header.Hash()
	return nil
}

// GetAncestorAtSlot gets the block row ancestor at a certain slot.
func (br *BlockRow) GetAncestorAtSlot(slot uint64) *BlockRow {
	if br.Header.Slot < slot {
		return nil
	}

	current := br

	// go up to the slot after the slot we're searching for
	for slot < current.Header.Slot {
		current = current.Parent
	}
	return current
}

var zeroHash = chainhash.Hash{}

func (br *BlockRow) attach(index *BlockIndex) error {
	if !br.parentHash.IsEqual(&zeroHash) {
		parentRow, found := index.get(br.parentHash)
		if !found {
			return fmt.Errorf("could not find parent in block index")
		}

		br.Parent = parentRow
	}
	return nil
}

// BlockIndex is an index from hash to BlockRow.
type BlockIndex struct {
	lock  sync.RWMutex
	index map[chainhash.Hash]*BlockRow
}

// Serialize serializes the block row to the specified writer.
func (i *BlockIndex) Serialize(w io.Writer) error {
	i.lock.RLock()
	defer i.lock.RUnlock()
	err := serializer.WriteVarInt(w, uint64(len(i.index)))
	if err != nil {
		return err
	}
	for _, row := range i.index {
		err = row.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Deserialize deserializes the block index from the specified reader.
func (i *BlockIndex) Deserialize(r io.Reader) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	buf, _ := r.(*bytes.Buffer)
	if buf.Len() > 0 {
		count, err := serializer.ReadVarInt(r)
		if err != nil {
			return err
		}
		i.index = make(map[chainhash.Hash]*BlockRow, count)
		for k := uint64(0); k < count; k++ {
			var row BlockRow
			err = row.Deserialize(r)
			if err != nil {
				return err
			}
			err = i.add(&row)
			if err != nil {
				return err
			}
		}
	}
	for _, r := range i.index {
		err := r.attach(i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *BlockIndex) get(hash chainhash.Hash) (*BlockRow, bool) {
	row, found := i.index[hash]
	return row, found
}

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
	blockHash := row.Header.Hash()
	i.index[blockHash] = row
	return nil
}

// SetDiskLocation sets the disk location of a block.
func (i *BlockIndex) SetDiskLocation(of chainhash.Hash, loc blockdb.BlockLocation) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	d, found := i.index[of]
	if !found {
		return fmt.Errorf("could not find block with hash %s", of)
	}

	d.Locator = loc

	i.index[of] = d
	return nil
}

// Add adds a row to the block index.
func (i *BlockIndex) Add(header primitives.BlockHeader) (*BlockRow, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	prev, found := i.index[header.PrevBlockHash]
	if !found {
		return nil, fmt.Errorf("could not add block to index: could not find parent with hash %s", header.PrevBlockHash)
	}

	row := &BlockRow{
		Header: header,
		Height: prev.Height + 1,
		Parent: prev,
		Hash:   header.Hash(),
	}

	err := i.add(row)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// InitBlocksIndex creates a new block index.
func InitBlocksIndex(genesisHeader primitives.BlockHeader, genesisLoc blockdb.BlockLocation) (*BlockIndex, error) {
	headerHash := genesisHeader.Hash()
	return &BlockIndex{
		index: map[chainhash.Hash]*BlockRow{
			headerHash: {
				Header:  genesisHeader,
				Locator: genesisLoc,
				Height:  0,
				Parent:  nil,
				Hash:    genesisHeader.Hash(),
			},
		},
	}, nil
}
