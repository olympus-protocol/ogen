package blockdb

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var (
	tipKey      = []byte("tip")
	finHeadKey  = []byte("finalized_head")
	jusHeadKey  = []byte("justified_head")
	finStateKey = []byte("finalized_state")
	jusStateKey = []byte("justified_state")
	genTimeKey  = []byte("genesis_key")

	blockRowPrefix = []byte("block-row-")
)

type Database interface {
	Close() error
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

var _ Database = &levelDB{}

type levelDB struct {
	log logger.Logger

	db *leveldb.DB

	canClose sync.WaitGroup
}

// NewLevelDB returns a database instance for storing blocks.
func NewLevelDB() (Database, error) {
	log := config.GlobalParams.Logger
	datapath := config.GlobalFlags.DataPath

	opts := &opt.Options{
		ErrorIfExist:           false,
		Strict:                 opt.DefaultStrict,
		Compression:            opt.SnappyCompression,
		Filter:                 filter.NewBloomFilter(10),
		DisableSeeksCompaction: true,
	}

	db, err := leveldb.OpenFile(datapath+"/chain", opts)
	if err != nil {
		return nil, err
	}

	blockdb := &levelDB{
		db:  db,
		log: log,
	}

	return blockdb, nil
}

// Close closes the database.
func (db *levelDB) Close() error {
	db.canClose.Wait()
	err := db.db.Close()
	if err != nil {
		return err
	}
	return nil
}

// GetBlock gets a block from the database.
func (db *levelDB) GetBlock(hash chainhash.Hash) (*primitives.Block, error) {
	blockBytes, err := db.get(hash[:])
	if err != nil {
		return nil, err
	}

	block := new(primitives.Block)
	err = block.Unmarshal(blockBytes)
	return block, err
}

// GetRawBlock gets a block serialized from the database.
func (db *levelDB) GetRawBlock(hash chainhash.Hash) ([]byte, error) {
	blockBytes, err := db.get(hash[:])
	if err != nil {
		return nil, err
	}

	return blockBytes, err
}

// AddRawBlock adds a raw block to the database.
func (db *levelDB) AddRawBlock(block *primitives.Block) error {
	blockHash := block.Hash()
	blockBytes, err := block.Marshal()
	if err != nil {
		return err
	}
	return db.set(blockHash[:], blockBytes)
}

// SetTip sets the current best tip of the blockchain.
func (db *levelDB) SetTip(c chainhash.Hash) error {
	return db.set(tipKey, c[:])
}

// GetTip gets the current best tip of the blockchain.
func (db *levelDB) GetTip() (chainhash.Hash, error) {
	return db.getHash(tipKey)
}

// SetFinalizedState sets the finalized state of the blockchain.
func (db *levelDB) SetFinalizedState(s state.State) error {
	b, err := s.Marshal()
	if err != nil {
		return err
	}

	return db.set(finStateKey, b)
}

// GetFinalizedState gets the finalized state of the blockchain.
func (db *levelDB) GetFinalizedState() (state.State, error) {

	stateBytes, err := db.get(finStateKey)
	if err != nil {
		return nil, err
	}

	s := state.NewEmptyState()

	err = s.Unmarshal(stateBytes)

	return s, err
}

// SetJustifiedState sets the justified state of the blockchain.
func (db *levelDB) SetJustifiedState(s state.State) error {
	b, err := s.Marshal()
	if err != nil {
		return err
	}

	return db.set(jusStateKey, b)
}

// GetJustifiedState gets the justified state of the blockchain.
func (db *levelDB) GetJustifiedState() (state.State, error) {

	stateBytes, err := db.get(jusStateKey)
	if err != nil {
		return nil, err
	}

	s := state.NewEmptyState()

	err = s.Unmarshal(stateBytes)

	return s, err
}

// SetBlockRow sets a block row on disk to store the block index.
func (db *levelDB) SetBlockRow(disk *primitives.BlockNodeDisk) error {
	key := append(blockRowPrefix, disk.Hash[:]...)
	b, err := disk.Marshal()
	if err != nil {
		return err
	}
	return db.set(key, b)
}

// GetBlockRow gets the block row on disk.
func (db *levelDB) GetBlockRow(c chainhash.Hash) (*primitives.BlockNodeDisk, error) {

	key := append(blockRowPrefix, c[:]...)

	b, err := db.get(key)
	if err != nil {
		return nil, err
	}

	r := new(primitives.BlockNodeDisk)

	err = r.Unmarshal(b)

	return r, err
}

// SetJustifiedHead sets the latest justified head.
func (db *levelDB) SetJustifiedHead(c chainhash.Hash) error {
	return db.set(jusHeadKey, c[:])
}

// GetJustifiedHead gets the latest justified head.
func (db *levelDB) GetJustifiedHead() (chainhash.Hash, error) {
	return db.getHash(jusHeadKey)
}

// SetFinalizedHead sets the finalized head of the blockchain.
func (db *levelDB) SetFinalizedHead(c chainhash.Hash) error {
	return db.set(finHeadKey, c[:])
}

// GetFinalizedHead gets the finalized head of the blockchain.
func (db *levelDB) GetFinalizedHead() (chainhash.Hash, error) {
	return db.getHash(finHeadKey)
}

// SetGenesisTime sets the genesis time of the blockchain.
func (db *levelDB) SetGenesisTime(t time.Time) error {
	bs, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	return db.set(genTimeKey, bs)
}

// GetGenesisTime gets the genesis time of the blockchain.
func (db *levelDB) GetGenesisTime() (time.Time, error) {

	bs, err := db.get(genTimeKey)
	if err != nil {
		return time.Time{}, err
	}

	var t time.Time
	err = t.UnmarshalBinary(bs)
	return t, err
}

func (db *levelDB) get(key []byte) ([]byte, error) {

	db.canClose.Add(1)
	defer db.canClose.Done()

	out, err := db.db.Get(key[:], nil)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (db *levelDB) getHash(key []byte) (chainhash.Hash, error) {

	db.canClose.Add(1)
	defer db.canClose.Done()

	var out chainhash.Hash
	b, err := db.db.Get(key[:], nil)
	if err != nil {
		return chainhash.Hash{}, err
	}
	copy(out[:], b)

	return out, nil
}

func (db *levelDB) set(k []byte, v []byte) error {

	db.canClose.Add(1)
	defer db.canClose.Done()

	err := db.db.Put(k, v, nil)
	if err != nil {
		return err
	}

	return nil
}
