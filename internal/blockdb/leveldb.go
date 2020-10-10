package blockdb

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

type levelDB struct {
	log       logger.Logger
	netParams *params.ChainParams

	lock     sync.Mutex
	db       *leveldb.DB
	canClose sync.WaitGroup
}

// NewLevelDB returns a database instance with a rawBlockDatabase and BadgerDB to use on the selected path.
func NewLevelDB() (Database, error) {
	datapath := config.GlobalFlags.DataPath
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	db, err := leveldb.OpenFile(datapath+"/chain", nil)
	if err != nil {
		return nil, err
	}
	blockdb := &levelDB{
		log:       log,
		db:        db,
		netParams: netParams,
	}
	return blockdb, nil
}

// Close closes the database.
func (db *levelDB) Close() {
	db.canClose.Wait()
	_ = db.db.Close()
}

// GetBlock gets a block from the database.
func (db *levelDB) GetBlock(hash chainhash.Hash) (*primitives.Block, error) {
	blockBytes, err := db.getKey(hash[:])
	if err != nil {
		return nil, err
	}

	block := new(primitives.Block)
	err = block.Unmarshal(blockBytes)
	return block, err
}

// GetRawBlock gets a block serialized from the database.
func (db *levelDB) GetRawBlock(hash chainhash.Hash) ([]byte, error) {
	blockBytes, err := db.getKey(hash[:])
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
	return db.setKey(blockHash[:], blockBytes)
}

// SetTip sets the current best tip of the blockchain.
func (db *levelDB) SetTip(c chainhash.Hash) error {
	return db.setKeyHash(tipKey, c)
}

// GetTip gets the current best tip of the blockchain.
func (db *levelDB) GetTip() (chainhash.Hash, error) {
	return db.getKeyHash(tipKey)
}

// SetFinalizedState sets the finalized state of the blockchain.
func (db *levelDB) SetFinalizedState(s state.State) error {
	buf, err := s.Marshal()
	if err != nil {
		return err
	}

	return db.setKey(finStateKey, buf)
}

// GetFinalizedState gets the finalized state of the blockchain.
func (db *levelDB) GetFinalizedState() (state.State, error) {
	stateBytes, err := db.getKey(finStateKey)
	if err != nil {
		return nil, err
	}
	s := state.NewEmptyState()
	err = s.Unmarshal(stateBytes)
	return s, err
}

// SetJustifiedState sets the justified state of the blockchain.
func (db *levelDB) SetJustifiedState(s state.State) error {
	buf, err := s.Marshal()
	if err != nil {
		return err
	}

	return db.setKey(jusStateKey, buf)
}

// GetJustifiedState gets the justified state of the blockchain.
func (db *levelDB) GetJustifiedState() (state.State, error) {
	stateBytes, err := db.getKey(jusStateKey)
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
	diskSer, err := disk.Marshal()
	if err != nil {
		return err
	}
	return db.setKey(key, diskSer)
}

// GetBlockRow gets the block row on disk.
func (db *levelDB) GetBlockRow(c chainhash.Hash) (*primitives.BlockNodeDisk, error) {
	key := append(blockRowPrefix, c[:]...)
	diskSer, err := db.getKey(key)
	if err != nil {
		return nil, err
	}

	d := new(primitives.BlockNodeDisk)
	err = d.Unmarshal(diskSer)
	return d, err
}

// SetJustifiedHead sets the latest justified head.
func (db *levelDB) SetJustifiedHead(c chainhash.Hash) error {
	return db.setKeyHash(jusHeadKey, c)
}

// GetJustifiedHead gets the latest justified head.
func (db *levelDB) GetJustifiedHead() (chainhash.Hash, error) {
	return db.getKeyHash(jusHeadKey)
}

// SetFinalizedHead sets the finalized head of the blockchain.
func (db *levelDB) SetFinalizedHead(c chainhash.Hash) error {
	return db.setKeyHash(finHeadKey, c)
}

// GetFinalizedHead gets the finalized head of the blockchain.
func (db *levelDB) GetFinalizedHead() (chainhash.Hash, error) {
	return db.getKeyHash(finHeadKey)
}

// SetGenesisTime sets the genesis time of the blockchain.
func (db *levelDB) SetGenesisTime(t time.Time) error {
	bs, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	return db.setKey(genTimeKey, bs)
}

// GetGenesisTime gets the genesis time of the blockchain.
func (db *levelDB) GetGenesisTime() (time.Time, error) {
	bs, err := db.getKey(genTimeKey)
	if err != nil {
		return time.Time{}, err
	}

	var t time.Time
	err = t.UnmarshalBinary(bs)
	return t, err
}

func (db *levelDB) getKeyHash(key []byte) (chainhash.Hash, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.canClose.Add(1)
	defer db.canClose.Done()
	var out chainhash.Hash
	h, err := db.db.Get(key, nil)
	if err != nil {
		return chainhash.Hash{}, err
	}
	copy(out[:], h)
	return out, nil
}

func (db *levelDB) getKey(key []byte) ([]byte, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.canClose.Add(1)
	defer db.canClose.Done()
	out, err := db.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (db *levelDB) setKeyHash(key []byte, to chainhash.Hash) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.canClose.Add(1)
	defer db.canClose.Done()
	err := db.db.Put(key, to[:], nil)
	if err != nil {
		return err
	}
	return nil
}

func (db *levelDB) setKey(key []byte, to []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.canClose.Add(1)
	defer db.canClose.Done()
	err := db.db.Put(key, to, nil)
	if err != nil {
		return err
	}
	return nil
}
