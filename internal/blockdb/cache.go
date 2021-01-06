package blockdb

import (
	"errors"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"sync"
	"time"
)

var ErrorNotInCache = errors.New("not in cache")

type Cache struct {
	cacheLock sync.Mutex
	cache     map[[32]byte][]byte
	count     int
}

// NewCacheDB returns a database instance for storing blocks.
func NewCacheDB() *Cache {

	c := &Cache{
		cache: make(map[[32]byte][]byte),
		count: 0,
	}

	return c

}

func (db *Cache) Flush() map[[32]byte][]byte {
	db.cacheLock.Lock()
	defer db.cacheLock.Unlock()

	cache := db.cache

	db.cache = make(map[[32]byte][]byte)
	db.count = 0

	return cache
}

// Count returns the amount of elements stored on cache
func (db *Cache) Count() int {
	return db.count
}
// GetBlock gets a block from the database.
func (db *Cache) GetBlock(hash chainhash.Hash) (*primitives.Block, error) {
	blockBytes, err := db.get(hash[:])
	if err != nil {
		return nil, err
	}

	block := new(primitives.Block)
	err = block.Unmarshal(blockBytes)
	return block, err
}

// GetRawBlock gets a block serialized from the database.
func (db *Cache) GetRawBlock(hash chainhash.Hash) ([]byte, error) {
	blockBytes, err := db.get(hash[:])
	if err != nil {
		return nil, err
	}
	return blockBytes, err
}

// AddRawBlock adds a raw block to the database.
func (db *Cache) AddRawBlock(block *primitives.Block) error {
	blockHash := block.Hash()
	blockBytes, err := block.Marshal()
	if err != nil {
		return err
	}
	return db.set(blockHash[:], blockBytes)
}

// SetTip sets the current best tip of the blockchain.
func (db *Cache) SetTip(c chainhash.Hash) error {
	return db.set(tipKey, c[:])
}

// GetTip gets the current best tip of the blockchain.
func (db *Cache) GetTip() (chainhash.Hash, error) {
	return db.getHash(tipKey)
}

// SetFinalizedState sets the finalized state of the blockchain.
func (db *Cache) SetFinalizedState(s state.State) error {
	buf, err := s.Marshal()
	if err != nil {
		return err
	}

	return db.set(finStateKey, buf)
}

// GetFinalizedState gets the finalized state of the blockchain.
func (db *Cache) GetFinalizedState() (state.State, error) {
	stateBytes, err := db.get(finStateKey)
	if err != nil {
		return nil, err
	}
	s := state.NewEmptyState()
	err = s.Unmarshal(stateBytes)
	return s, err
}

// SetJustifiedState sets the justified state of the blockchain.
func (db *Cache) SetJustifiedState(s state.State) error {
	buf, err := s.Marshal()
	if err != nil {
		return err
	}

	return db.set(jusStateKey, buf)
}

// GetJustifiedState gets the justified state of the blockchain.
func (db *Cache) GetJustifiedState() (state.State, error) {
	stateBytes, err := db.get(jusStateKey)
	if err != nil {
		return nil, err
	}
	s := state.NewEmptyState()
	err = s.Unmarshal(stateBytes)
	return s, err
}

// SetBlockRow sets a block row on disk to store the block index.
func (db *Cache) SetBlockRow(disk *primitives.BlockNodeDisk) error {
	key := append(blockRowPrefix, disk.Hash[:]...)
	diskSer, err := disk.Marshal()
	if err != nil {
		return err
	}
	return db.set(key, diskSer)
}

// GetBlockRow gets the block row on disk.
func (db *Cache) GetBlockRow(c chainhash.Hash) (*primitives.BlockNodeDisk, error) {
	key := append(blockRowPrefix, c[:]...)
	diskSer, err := db.get(key)
	if err != nil {
		return nil, err
	}

	d := new(primitives.BlockNodeDisk)
	err = d.Unmarshal(diskSer)
	return d, err
}

// SetJustifiedHead sets the latest justified head.
func (db *Cache) SetJustifiedHead(c chainhash.Hash) error {
	return db.set(jusHeadKey, c[:])
}

// GetJustifiedHead gets the latest justified head.
func (db *Cache) GetJustifiedHead() (chainhash.Hash, error) {
	return db.getHash(jusHeadKey)
}

// SetFinalizedHead sets the finalized head of the blockchain.
func (db *Cache) SetFinalizedHead(c chainhash.Hash) error {
	return db.set(finHeadKey, c[:])
}

// GetFinalizedHead gets the finalized head of the blockchain.
func (db *Cache) GetFinalizedHead() (chainhash.Hash, error) {
	return db.getHash(finHeadKey)
}

// SetGenesisTime sets the genesis time of the blockchain.
func (db *Cache) SetGenesisTime(t time.Time) error {
	bs, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	return db.set(genTimeKey, bs)
}

// GetGenesisTime gets the genesis time of the blockchain.
func (db *Cache) GetGenesisTime() (time.Time, error) {
	bs, err := db.get(genTimeKey)
	if err != nil {
		return time.Time{}, err
	}

	var t time.Time
	err = t.UnmarshalBinary(bs)
	return t, err
}

func (db *Cache) getHash(key []byte) (chainhash.Hash, error) {
	db.cacheLock.Lock()
	defer db.cacheLock.Unlock()

	var k [32]byte
	copy(k[:], key)

	out, ok := db.cache[k]
	if !ok {
		return chainhash.Hash{}, ErrorNotInCache
	}

	var hash chainhash.Hash
	copy(hash[:], out)

	return hash, nil
}

func (db *Cache) get(key []byte) ([]byte, error) {
	db.cacheLock.Lock()
	defer db.cacheLock.Unlock()

	var k [32]byte
	copy(k[:], key)

	out, ok := db.cache[k]
	if !ok {
		return nil, ErrorNotInCache
	}

	return out, nil
}

func (db *Cache) set(key []byte, to []byte) error {
	db.cacheLock.Lock()
	defer db.cacheLock.Unlock()

	var k [32]byte
	copy(k[:], key)
	db.cache[k] = to
	db.count += 1

	return nil
}
