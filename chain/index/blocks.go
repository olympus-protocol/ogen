package index

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// BlockRow represents a single row in the block index.
type BlockRow struct {
	Header     p2p.BlockHeader
	Locator    blockdb.BlockLocation
	Height     int32
	parentHash chainhash.Hash
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
		parentHash, err = br.Parent.Header.Hash()
		if err != nil {
			return err
		}
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
	return nil
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
	blockHash, err := row.Header.Hash()
	if err != nil {
		return err
	}
	i.index[blockHash] = row
	return nil
}

// Add adds a row to the block index.
func (i *BlockIndex) Add(header p2p.BlockHeader, locator blockdb.BlockLocation) (*BlockRow, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	prev, found := i.index[header.PrevBlockHash]
	if !found {
		return nil, fmt.Errorf("could not add block to index: could not find parent with hash %s", header.PrevBlockHash)
	}

	row := &BlockRow{
		Header:  header,
		Locator: locator,
		Height:  prev.Height + 1,
		Parent:  prev,
	}

	err := i.add(row)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// InitBlocksIndex creates a new block index.
func InitBlocksIndex(genesisHeader p2p.BlockHeader, genesisLoc blockdb.BlockLocation) (*BlockIndex, error) {
	headerHash, err := genesisHeader.Hash()
	if err != nil {
		return nil, err
	}
	return &BlockIndex{
		index: map[chainhash.Hash]*BlockRow{
			headerHash: {
				Header:  genesisHeader,
				Locator: genesisLoc,
				Height:  0,
				Parent:  nil,
			},
		},
	}, nil
}
