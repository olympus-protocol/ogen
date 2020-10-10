package blockdb

import (
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"time"
)

type BlockDB interface {
	Close()
	GetBlock(hash chainhash.Hash) (*primitives.Block, error)
	GetRawBlock(hash chainhash.Hash) ([]byte, error)
	AddRawBlock(block *primitives.Block) error
	SetTip(c chainhash.Hash) error
	GetTip() (chainhash.Hash, error)
	SetFinalizedState(s state.State) error
	GetFinalizedState() (state.State, error)
	SetJustifiedState(s state.State) error
	GetJustifiedState() (state.State, error)
	SetBlockRow(disk *primitives.BlockNodeDisk) error
	GetBlockRow(c chainhash.Hash) (*primitives.BlockNodeDisk, error)
	SetJustifiedHead(c chainhash.Hash) error
	GetJustifiedHead() (chainhash.Hash, error)
	SetFinalizedHead(c chainhash.Hash) error
	GetFinalizedHead() (chainhash.Hash, error)
	SetGenesisTime(t time.Time) error
	GetGenesisTime() (time.Time, error)
}

type blockDB struct {
	sync   bool
	memory Database
	disk   Database
}

func (b blockDB) Close() {
	b.memory.Close()
	b.disk.Close()
}

func (b blockDB) GetBlock(hash chainhash.Hash) (*primitives.Block, error) {
	return b.disk.GetBlock(hash)
}

func (b blockDB) GetRawBlock(hash chainhash.Hash) ([]byte, error) {
	return b.disk.GetRawBlock(hash)
}

func (b blockDB) AddRawBlock(block *primitives.Block) error {
	// TODO switch between memory and disk
	return nil
}

func (b blockDB) SetTip(c chainhash.Hash) error {
	// TODO switch between memory and disk
	return nil
}

func (b blockDB) GetTip() (chainhash.Hash, error) {
	return b.disk.GetTip()
}

func (b blockDB) SetFinalizedState(s state.State) error {
	panic("implement me")
}

func (b blockDB) GetFinalizedState() (state.State, error) {
	return b.disk.GetFinalizedState()
}

func (b blockDB) SetJustifiedState(s state.State) error {
	panic("implement me")
}

func (b blockDB) GetJustifiedState() (state.State, error) {
	return b.disk.GetJustifiedState()
}

func (b blockDB) SetBlockRow(disk *primitives.BlockNodeDisk) error {
	panic("implement me")
}

func (b blockDB) GetBlockRow(c chainhash.Hash) (*primitives.BlockNodeDisk, error) {
	return b.disk.GetBlockRow(c)
}

func (b blockDB) SetJustifiedHead(c chainhash.Hash) error {
	panic("implement me")
}

func (b blockDB) GetJustifiedHead() (chainhash.Hash, error) {
	return b.disk.GetJustifiedHead()
}

func (b blockDB) SetFinalizedHead(c chainhash.Hash) error {
	panic("implement me")
}

func (b blockDB) GetFinalizedHead() (chainhash.Hash, error) {
	return b.disk.GetFinalizedHead()
}

func (b blockDB) SetGenesisTime(t time.Time) error {
	panic("implement me")
}

func (b blockDB) GetGenesisTime() (time.Time, error) {
	return b.disk.GetGenesisTime()
}

func NewBlockDB() (BlockDB, error) {
	memdb, err := NewMemoryDB()
	if err != nil {
		return nil, err
	}
	diskdb, err := NewBoltDB()
	if err != nil {
		return nil, err
	}
	return &blockDB{
		sync:   true,
		memory: memdb,
		disk:   diskdb,
	}, nil
}
