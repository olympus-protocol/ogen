package mock

import (
	"bytes"
	"errors"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"sync"
)

type MemoryDB struct {
	lock sync.Mutex
	blocks [][]byte
}

func (MemoryDB) Close() {}

func (m *MemoryDB) GetRawBlock(locator blockdb.BlockLocation, hash chainhash.Hash) ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if uint32(len(m.blocks)) <= locator.FileNum {
		return nil, errors.New("invalid block locator")
	}

	return m.blocks[locator.FileNum], nil
}

func (m *MemoryDB) AddRawBlock(block *primitives.Block) (*blockdb.BlockLocation, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	blockBuf := bytes.NewBuffer([]byte{})
	_ = block.Encode(blockBuf)
	m.blocks = append(m.blocks, blockBuf.Bytes())
	return &blockdb.BlockLocation{
		FileNum:     uint32(len(m.blocks) - 1),
		BlockOffset: 0,
		BlockSize:   0,
	}, nil
}

func (m *MemoryDB) Clear() {
	panic("implement me")
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		blocks: make([][]byte, 0, 1000),
	}
}