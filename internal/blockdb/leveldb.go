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
	Commit() error
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

	cache    *Cache
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
		db:    db,
		cache: NewCacheDB(),
		log:   log,
	}

	go blockdb.committer()

	return blockdb, nil
}

func (db *levelDB) committer() {
	for {
		time.Sleep(time.Second * 5)
		if db.cache.Count() > 5000 {
			err := db.Commit()
			if err != nil {
				db.log.Error(err)
			}
		}
	}
}

// Commit grabs all information on the cache and stores on disk
func (db *levelDB) Commit() error {
	m := db.cache.Flush()
	batch := leveldb.MakeBatch(len(m))
	for k, v := range m {
		batch.Put(k[:], v)
	}
	err := db.db.Write(batch, nil)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the database.
func (db *levelDB) Close() error {
	db.canClose.Wait()
	err := db.Commit()
	if err != nil {
		return err
	}
	err = db.db.Close()
	if err != nil {
		return err
	}
	return nil
}

// GetBlock gets a block from the database.
func (db *levelDB) GetBlock(hash chainhash.Hash) (*primitives.Block, error) {
	b, err := db.cache.GetBlock(hash)
	if err == nil {
		return b, nil
	}

	blockBytes, err := db.get(hash)
	if err != nil {
		return nil, err
	}

	block := new(primitives.Block)
	err = block.Unmarshal(blockBytes)
	return block, err
}

// GetRawBlock gets a block serialized from the database.
func (db *levelDB) GetRawBlock(hash chainhash.Hash) ([]byte, error) {
	b, err := db.cache.GetRawBlock(hash)
	if err == nil {
		return b, nil
	}

	blockBytes, err := db.get(hash)
	if err != nil {
		return nil, err
	}

	return blockBytes, err
}

// AddRawBlock adds a raw block to the database.
func (db *levelDB) AddRawBlock(block *primitives.Block) error {
	return db.cache.AddRawBlock(block)
}

// SetTip sets the current best tip of the blockchain.
func (db *levelDB) SetTip(c chainhash.Hash) error {
	return db.cache.SetTip(c)
}

// GetTip gets the current best tip of the blockchain.
func (db *levelDB) GetTip() (chainhash.Hash, error) {
	t, err := db.cache.GetTip()
	if err == nil {
		return t, nil
	}

	var k [32]byte
	copy(k[:], tipKey)
	return db.getHash(k)
}

// SetFinalizedState sets the finalized state of the blockchain.
func (db *levelDB) SetFinalizedState(s state.State) error {
	return db.cache.SetFinalizedState(s)
}

// GetFinalizedState gets the finalized state of the blockchain.
func (db *levelDB) GetFinalizedState() (state.State, error) {
	s, err := db.cache.GetFinalizedState()
	if err == nil {
		return s, nil
	}

	var k [32]byte
	copy(k[:], finStateKey)

	stateBytes, err := db.get(k)
	if err != nil {
		return nil, err
	}

	s = state.NewEmptyState()

	err = s.Unmarshal(stateBytes)

	return s, err
}

// SetJustifiedState sets the justified state of the blockchain.
func (db *levelDB) SetJustifiedState(s state.State) error {
	return db.cache.SetJustifiedState(s)
}

// GetJustifiedState gets the justified state of the blockchain.
func (db *levelDB) GetJustifiedState() (state.State, error) {
	s, err := db.cache.GetJustifiedState()
	if err == nil {
		return s, nil
	}

	var k [32]byte
	copy(k[:], jusStateKey)

	stateBytes, err := db.get(k)
	if err != nil {
		return nil, err
	}

	s = state.NewEmptyState()

	err = s.Unmarshal(stateBytes)

	return s, err
}

// SetBlockRow sets a block row on disk to store the block index.
func (db *levelDB) SetBlockRow(disk *primitives.BlockNodeDisk) error {
	return db.cache.SetBlockRow(disk)
}

// GetBlockRow gets the block row on disk.
func (db *levelDB) GetBlockRow(c chainhash.Hash) (*primitives.BlockNodeDisk, error) {
	r, err := db.cache.GetBlockRow(c)
	if err == nil {
		return r, nil
	}

	key := append(blockRowPrefix, c[:]...)

	var k [32]byte
	copy(k[:], key)

	b, err := db.get(k)
	if err != nil {
		return nil, err
	}

	r = new(primitives.BlockNodeDisk)

	err = r.Unmarshal(b)

	return r, err
}

// SetJustifiedHead sets the latest justified head.
func (db *levelDB) SetJustifiedHead(c chainhash.Hash) error {
	return db.cache.SetJustifiedHead(c)
}

// GetJustifiedHead gets the latest justified head.
func (db *levelDB) GetJustifiedHead() (chainhash.Hash, error) {
	h, err := db.cache.GetJustifiedHead()
	if err == nil {
		return h, nil
	}

	var k [32]byte
	copy(k[:], jusHeadKey)

	return db.getHash(k)
}

// SetFinalizedHead sets the finalized head of the blockchain.
func (db *levelDB) SetFinalizedHead(c chainhash.Hash) error {
	return db.cache.SetFinalizedHead(c)
}

// GetFinalizedHead gets the finalized head of the blockchain.
func (db *levelDB) GetFinalizedHead() (chainhash.Hash, error) {
	h, err := db.cache.GetFinalizedHead()
	if err == nil {
		return h, nil
	}

	var k [32]byte
	copy(k[:], finHeadKey)

	return db.getHash(k)
}

// SetGenesisTime sets the genesis time of the blockchain.
func (db *levelDB) SetGenesisTime(t time.Time) error {
	return db.cache.SetGenesisTime(t)
}

// GetGenesisTime gets the genesis time of the blockchain.
func (db *levelDB) GetGenesisTime() (time.Time, error) {
	t, err := db.cache.GetGenesisTime()
	if err == nil {
		return t, nil
	}

	var k [32]byte
	copy(k[:], genTimeKey)

	bs, err := db.get(k)
	if err != nil {
		return time.Time{}, err
	}

	err = t.UnmarshalBinary(bs)
	return t, err
}

func (db *levelDB) get(key [32]byte) ([]byte, error) {

	db.canClose.Add(1)
	defer db.canClose.Done()

	out, err := db.db.Get(key[:], nil)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (db *levelDB) getHash(key [32]byte) (chainhash.Hash, error) {

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