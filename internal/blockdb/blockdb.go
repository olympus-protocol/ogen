package blockdb

import (
	"fmt"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger"
)

// BlockDB is an interface for blockDb
type BlockDB interface {
	Close()
	Update(cb func(txn DBUpdateTransaction) error) error
	View(cb func(txn DBViewTransaction) error) error
}

type blockDB struct {
	log      logger.Logger
	badgerdb *badger.DB
	params   params.ChainParams
	lock     sync.RWMutex

	requestedClose uint32
	canClose       sync.WaitGroup
}

var _ BlockDB = &blockDB{}


type UpdateTransaction struct {
	ReadTransaction
}

type ReadTransaction struct {
	db          *blockDB
	log         logger.Logger
	transaction *badger.Txn
}

// NewBlockDB returns a database instance with a rawBlockDatabase and BadgerDB to use on the selected path.
func NewBlockDB(path string, params params.ChainParams, log logger.Logger) (BlockDB, error) {
	badgerdb, err := badger.Open(badger.DefaultOptions(path + "/chain").WithLogger(nil))
	if err != nil {
		return nil, err
	}
	blockdb := &blockDB{
		log:      log,
		badgerdb: badgerdb,
		params:   params,
	}
	return blockdb, nil
}

// Close closes the database.
func (bdb *blockDB) Close() {
	if atomic.LoadUint32(&bdb.requestedClose) != 0 {
		return
	}
	atomic.StoreUint32(&bdb.requestedClose, 1)
	bdb.canClose.Wait()
	_ = bdb.badgerdb.Close()
}

// Update gets a transaction for updating the database.
func (bdb *blockDB) Update(cb func(txn DBUpdateTransaction) error) error {
	bdb.lock.Lock()
	defer bdb.lock.Unlock()
	if atomic.LoadUint32(&bdb.requestedClose) != 0 {
		return fmt.Errorf("database is closing")
	}

	bdb.canClose.Add(1)
	defer bdb.canClose.Done()
	return bdb.badgerdb.Update(func(tx *badger.Txn) error {
		blockTxn := ReadTransaction{
			db:          bdb,
			log:         bdb.log,
			transaction: tx,
		}

		return cb(&UpdateTransaction{blockTxn})
	})
}

// View gets a transaction for viewing the database.
func (bdb *blockDB) View(cb func(txn DBViewTransaction) error) error {
	bdb.lock.RLock()
	defer bdb.lock.RUnlock()
	if atomic.LoadUint32(&bdb.requestedClose) != 0 {
		return fmt.Errorf("database is closing")
	}

	bdb.canClose.Add(1)
	defer bdb.canClose.Done()
	return bdb.badgerdb.Update(func(tx *badger.Txn) error {
		blockTxn := &ReadTransaction{
			db:          bdb,
			log:         bdb.log,
			transaction: tx,
		}

		return cb(blockTxn)
	})
}

// GetBlock gets a block from the database.
func (brt *ReadTransaction) GetBlock(hash chainhash.Hash) (*primitives.Block, error) {
	blockBytes, err := getKey(brt, hash[:])
	if err != nil {
		return nil, err
	}

	block := new(primitives.Block)
	err = block.Unmarshal(blockBytes)
	return block, err
}

// GetRawBlock gets a block serialized from the database.
func (brt *ReadTransaction) GetRawBlock(hash chainhash.Hash) ([]byte, error) {
	blockBytes, err := getKey(brt, hash[:])
	if err != nil {
		return nil, err
	}
	return blockBytes, err
}

// AddRawBlock adds a raw block to the database.
func (but *UpdateTransaction) AddRawBlock(block *primitives.Block) error {
	blockHash := block.Hash()
	blockBytes, err := block.Marshal()
	if err != nil {
		return err
	}
	return setKey(but, blockHash[:], blockBytes)
}

func getKeyHash(tx *ReadTransaction, key []byte) (chainhash.Hash, error) {
	var out chainhash.Hash
	i, err := tx.transaction.Get(key)
	if err != nil {
		return chainhash.Hash{}, err
	}
	_, err = i.ValueCopy(out[:])
	if err != nil {
		return out, err
	}
	return out, nil
}

func getKey(tx *ReadTransaction, key []byte) ([]byte, error) {
	var out []byte
	i, err := tx.transaction.Get(key)
	if err != nil {
		return nil, err
	}
	out, err = i.ValueCopy(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func setKeyHash(tx *UpdateTransaction, key []byte, to chainhash.Hash) error {
	return tx.transaction.Set(key, to[:])
}

func setKey(tx *UpdateTransaction, key []byte, to []byte) error {
	return tx.transaction.Set(key, to)
}

var tipKey = []byte("chain-tip")

// SetTip sets the current best tip of the blockchain.
func (but *UpdateTransaction) SetTip(c chainhash.Hash) error {
	return setKeyHash(but, tipKey, c)
}

// GetTip gets the current best tip of the blockchain.
func (brt *ReadTransaction) GetTip() (chainhash.Hash, error) {
	return getKeyHash(brt, tipKey)
}

var finalizedStateKey = []byte("finalized-state")

// SetFinalizedState sets the finalized state of the blockchain.
func (but *UpdateTransaction) SetFinalizedState(s state.State) error {
	buf, err := s.Marshal()
	if err != nil {
		return err
	}

	return setKey(but, finalizedStateKey, buf)
}

// GetFinalizedState gets the finalized state of the blockchain.
func (brt *ReadTransaction) GetFinalizedState() (state.State, error) {
	stateBytes, err := getKey(brt, finalizedStateKey)
	if err != nil {
		return nil, err
	}
	s := state.NewEmptyState()
	err = s.Unmarshal(stateBytes)
	return s, err
}

var justifiedStateKey = []byte("justified-state")

// SetJustifiedState sets the justified state of the blockchain.
func (but *UpdateTransaction) SetJustifiedState(s state.State) error {
	buf, err := s.Marshal()
	if err != nil {
		return err
	}

	return setKey(but, justifiedStateKey, buf)
}

// GetJustifiedState gets the justified state of the blockchain.
func (brt *ReadTransaction) GetJustifiedState() (state.State, error) {
	stateBytes, err := getKey(brt, justifiedStateKey)
	if err != nil {
		return nil, err
	}
	s := state.NewEmptyState()
	err = s.Unmarshal(stateBytes)
	return s, err
}

var blockRowPrefix = []byte("block-row")

// SetBlockRow sets a block row on disk to store the block index.
func (but *UpdateTransaction) SetBlockRow(disk *primitives.BlockNodeDisk) error {
	key := append(blockRowPrefix, disk.Hash[:]...)
	diskSer, err := disk.Marshal()
	if err != nil {
		return err
	}
	return setKey(but, key, diskSer)
}

// GetBlockRow gets the block row on disk.
func (brt *ReadTransaction) GetBlockRow(c chainhash.Hash) (*primitives.BlockNodeDisk, error) {
	key := append(blockRowPrefix, c[:]...)
	diskSer, err := getKey(brt, key)
	if err != nil {
		return nil, err
	}

	d := new(primitives.BlockNodeDisk)
	err = d.Unmarshal(diskSer)
	return d, err
}

var justifiedHeadKey = []byte("justified-head")

// SetJustifiedHead sets the latest justified head.
func (but *UpdateTransaction) SetJustifiedHead(c chainhash.Hash) error {
	return setKeyHash(but, justifiedHeadKey, c)
}

// GetJustifiedHead gets the latest justified head.
func (brt *ReadTransaction) GetJustifiedHead() (chainhash.Hash, error) {
	return getKeyHash(brt, justifiedHeadKey)
}

var finalizedHeadKey = []byte("finalized-head")

// SetFinalizedHead sets the finalized head of the blockchain.
func (but *UpdateTransaction) SetFinalizedHead(c chainhash.Hash) error {
	return setKeyHash(but, finalizedHeadKey, c)
}

// GetFinalizedHead gets the finalized head of the blockchain.
func (brt *ReadTransaction) GetFinalizedHead() (chainhash.Hash, error) {
	return getKeyHash(brt, finalizedHeadKey)
}

var genesisTimeKey = []byte("genesisTime")

// SetGenesisTime sets the genesis time of the blockchain.
func (but *UpdateTransaction) SetGenesisTime(t time.Time) error {
	bs, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	return setKey(but, genesisTimeKey, bs)
}

// GetGenesisTime gets the genesis time of the blockchain.
func (brt *ReadTransaction) GetGenesisTime() (time.Time, error) {
	bs, err := getKey(brt, genesisTimeKey)
	if err != nil {
		return time.Time{}, err
	}

	var t time.Time
	err = t.UnmarshalBinary(bs)
	return t, err
}

var _ DB = &blockDB{}
var _ DBUpdateTransaction = &UpdateTransaction{}
var _ DBViewTransaction = &ReadTransaction{}

// DB is a database for storing chain state.
type DB interface {
	Close()
	Update(func(DBUpdateTransaction) error) error
	View(func(DBViewTransaction) error) error
}

// DBTransactionRead is a transaction to view the state of the database.
type DBViewTransaction interface {
	GetBlock(hash chainhash.Hash) (*primitives.Block, error)
	GetRawBlock(hash chainhash.Hash) ([]byte, error)
	GetTip() (chainhash.Hash, error)
	GetFinalizedState() (state.State, error)
	GetJustifiedState() (state.State, error)
	GetBlockRow(chainhash.Hash) (*primitives.BlockNodeDisk, error)
	GetJustifiedHead() (chainhash.Hash, error)
	GetFinalizedHead() (chainhash.Hash, error)
	GetGenesisTime() (time.Time, error)
}

// DBTransaction is a transaction to update the state of the database.
type DBUpdateTransaction interface {
	AddRawBlock(block *primitives.Block) error
	SetTip(chainhash.Hash) error
	SetFinalizedState(state.State) error
	SetJustifiedState(state.State) error
	SetBlockRow(disk *primitives.BlockNodeDisk) error
	SetJustifiedHead(chainhash.Hash) error
	SetFinalizedHead(chainhash.Hash) error
	SetGenesisTime(time.Time) error
	DBViewTransaction
}
