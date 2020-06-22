package bdb

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/logger"

	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type BlockDB struct {
	log      *logger.Logger
	badgerdb *badger.DB
	params   params.ChainParams
	lock     sync.RWMutex

	requestedClose uint32
	canClose       sync.WaitGroup
}

type BlockDBUpdateTransaction struct {
	BlockDBReadTransaction
}

type BlockDBReadTransaction struct {
	db          *BlockDB
	log         *logger.Logger
	transaction *badger.Txn
}

// NewBlockDB returns a database instance with a rawBlockDatabase and BadgerDB to use on the selected path.
func NewBlockDB(path string, params params.ChainParams, log *logger.Logger) (*BlockDB, error) {
	badgerdb, err := badger.Open(badger.DefaultOptions(path + "/chain").WithLogger(nil))
	if err != nil {
		return nil, err
	}
	blockdb := &BlockDB{
		log:      log,
		badgerdb: badgerdb,
		params:   params,
	}
	return blockdb, nil
}

// Close closes the database.
func (bdb *BlockDB) Close() {
	if atomic.LoadUint32(&bdb.requestedClose) != 0 {
		return
	}
	atomic.StoreUint32(&bdb.requestedClose, 1)
	bdb.canClose.Wait()
	_ = bdb.badgerdb.Close()
}

// Update gets a transaction for updating the database.
func (bdb *BlockDB) Update(cb func(txn DBUpdateTransaction) error) error {
	bdb.lock.Lock()
	defer bdb.lock.Unlock()
	if atomic.LoadUint32(&bdb.requestedClose) != 0 {
		return fmt.Errorf("database is closing")
	}

	bdb.canClose.Add(1)
	defer bdb.canClose.Done()
	return bdb.badgerdb.Update(func(tx *badger.Txn) error {
		blockTxn := BlockDBReadTransaction{
			db:          bdb,
			log:         bdb.log,
			transaction: tx,
		}

		return cb(&BlockDBUpdateTransaction{blockTxn})
	})
}

// View gets a transaction for viewing the database.
func (bdb *BlockDB) View(cb func(txn DBViewTransaction) error) error {
	bdb.lock.RLock()
	defer bdb.lock.RUnlock()
	if atomic.LoadUint32(&bdb.requestedClose) != 0 {
		return fmt.Errorf("database is closing")
	}

	bdb.canClose.Add(1)
	defer bdb.canClose.Done()
	return bdb.badgerdb.Update(func(tx *badger.Txn) error {
		blockTxn := &BlockDBReadTransaction{
			db:          bdb,
			log:         bdb.log,
			transaction: tx,
		}

		return cb(blockTxn)
	})
}

// GetBlock gets a block from the database.
func (brt *BlockDBReadTransaction) GetBlock(hash chainhash.Hash) (*primitives.Block, error) {
	blockBytes, err := getKey(brt, hash[:])
	if err != nil {
		return nil, err
	}

	block := new(primitives.Block)
	err = block.Unmarshal(blockBytes)
	return block, err
}

// GetRawBlock gets a block serialized from the database.
func (brt *BlockDBReadTransaction) GetRawBlock(hash chainhash.Hash) ([]byte, error) {
	blockBytes, err := getKey(brt, hash[:])
	if err != nil {
		return nil, err
	}
	return blockBytes, err
}

// AddRawBlock adds a raw block to the database.
func (but *BlockDBUpdateTransaction) AddRawBlock(block *primitives.Block) error {
	blockHash, err := block.Hash()
	if err != nil {
		return err
	}
	blockBytes, err := block.Marshal()
	if err != nil {
		return err
	}
	return setKey(but, blockHash[:], blockBytes)
}

func getKeyHash(tx *BlockDBReadTransaction, key []byte) (chainhash.Hash, error) {
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

func getKey(tx *BlockDBReadTransaction, key []byte) ([]byte, error) {
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

func setKeyHash(tx *BlockDBUpdateTransaction, key []byte, to chainhash.Hash) error {
	return tx.transaction.Set(key, to[:])
}

func setKey(tx *BlockDBUpdateTransaction, key []byte, to []byte) error {
	return tx.transaction.Set(key, to)
}

var latestVotePrefix = []byte("latest-vote")

var tipKey = []byte("chain-tip")

// SetTip sets the current best tip of the blockchain.
func (but *BlockDBUpdateTransaction) SetTip(c chainhash.Hash) error {
	return setKeyHash(but, tipKey, c)
}

// GetTip gets the current best tip of the blockchain.
func (brt *BlockDBReadTransaction) GetTip() (chainhash.Hash, error) {
	return getKeyHash(brt, tipKey)
}

var finalizedStateKey = []byte("finalized-state")

// SetFinalizedState sets the finalized state of the blockchain.
func (but *BlockDBUpdateTransaction) SetFinalizedState(s *primitives.State) error {
	buf := bytes.NewBuffer([]byte{})
	if err := s.Serialize(buf); err != nil {
		return err
	}

	return setKey(but, finalizedStateKey, buf.Bytes())
}

// GetFinalizedState gets the finalized state of the blockchain.
func (brt *BlockDBReadTransaction) GetFinalizedState() (*primitives.State, error) {
	stateBytes, err := getKey(brt, finalizedStateKey)
	if err != nil {
		return nil, err
	}
	stateBuf := bytes.NewBuffer(stateBytes)
	state := new(primitives.State)
	err = state.Deserialize(stateBuf)
	return state, err
}

var justifiedStateKey = []byte("justified-state")

// SetJustifiedState sets the justified state of the blockchain.
func (but *BlockDBUpdateTransaction) SetJustifiedState(s *primitives.State) error {
	buf := bytes.NewBuffer([]byte{})
	if err := s.Serialize(buf); err != nil {
		return err
	}

	return setKey(but, justifiedStateKey, buf.Bytes())
}

// GetJustifiedState gets the justified state of the blockchain.
func (brt *BlockDBReadTransaction) GetJustifiedState() (*primitives.State, error) {
	stateBytes, err := getKey(brt, justifiedStateKey)
	if err != nil {
		return nil, err
	}
	stateBuf := bytes.NewBuffer(stateBytes)
	state := new(primitives.State)
	err = state.Deserialize(stateBuf)
	return state, err
}

var blockRowPrefix = []byte("block-row")

// SetBlockRow sets a block row on disk to store the block index.
func (but *BlockDBUpdateTransaction) SetBlockRow(disk *BlockNodeDisk) error {
	key := append(blockRowPrefix, disk.Hash[:]...)
	diskSer, err := disk.Marshal()
	if err != nil {
		return err
	}
	return setKey(but, key, diskSer)
}

// GetBlockRow gets the block row on disk.
func (brt *BlockDBReadTransaction) GetBlockRow(c chainhash.Hash) (*BlockNodeDisk, error) {
	key := append(blockRowPrefix, c[:]...)
	diskSer, err := getKey(brt, key)
	if err != nil {
		return nil, err
	}

	d := new(BlockNodeDisk)
	err = d.Unmarshal(diskSer)
	return d, err
}

var justifiedHeadKey = []byte("justified-head")

// SetJustifiedHead sets the latest justified head.
func (but *BlockDBUpdateTransaction) SetJustifiedHead(c chainhash.Hash) error {
	return setKeyHash(but, justifiedHeadKey, c)
}

// GetJustifiedHead gets the latest justified head.
func (brt *BlockDBReadTransaction) GetJustifiedHead() (chainhash.Hash, error) {
	return getKeyHash(brt, justifiedHeadKey)
}

var finalizedHeadKey = []byte("finalized-head")

// SetFinalizedHead sets the finalized head of the blockchain.
func (but *BlockDBUpdateTransaction) SetFinalizedHead(c chainhash.Hash) error {
	return setKeyHash(but, finalizedHeadKey, c)
}

// GetFinalizedHead gets the finalized head of the blockchain.
func (brt *BlockDBReadTransaction) GetFinalizedHead() (chainhash.Hash, error) {
	return getKeyHash(brt, finalizedHeadKey)
}

var genesisTimeKey = []byte("genesisTime")

// SetGenesisTime sets the genesis time of the blockchain.
func (but *BlockDBUpdateTransaction) SetGenesisTime(t time.Time) error {
	bs, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	return setKey(but, genesisTimeKey, bs)
}

// GetGenesisTime gets the genesis time of the blockchain.
func (brt *BlockDBReadTransaction) GetGenesisTime() (time.Time, error) {
	bs, err := getKey(brt, genesisTimeKey)
	if err != nil {
		return time.Time{}, err
	}

	var t time.Time
	err = t.UnmarshalBinary(bs)
	return t, err
}

var accountPrefix = []byte("account-")

// GetAccountTxs returns accounts transactions.
func (brt *BlockDBReadTransaction) GetAccountTxs(acc [20]byte) (*primitives.AccountTxs, error) {
	// TODO handle non existent accounts on index.
	key := append(accountPrefix, acc[:]...)
	accTxsBs, err := getKey(brt, key)
	if err != nil {
		return nil, err
	}
	accs := new(primitives.AccountTxs)
	buf := bytes.NewBuffer(accTxsBs)
	err = accs.Decode(buf)
	if err != nil {
		return nil, err
	}
	return accs, nil
}

// SetAccountTx adds a new tx hash to the account txs slice.
func (but *BlockDBUpdateTransaction) SetAccountTx(acc [20]byte, hash chainhash.Hash) error {
	// TODO handle non existent accounts on index.
	key := append(accountPrefix, acc[:]...)
	accTxsBs, err := getKey(&but.BlockDBReadTransaction, key)
	if err != nil {
		return err
	}
	accs := new(primitives.AccountTxs)
	buf := bytes.NewBuffer(accTxsBs)
	err = accs.Decode(buf)
	if err != nil {
		return err
	}
	accs.TxsAmount = +1
	accs.Txs = append(accs.Txs, hash)
	newBuf := bytes.NewBuffer([]byte{})
	err = accs.Encode(newBuf)
	if err != nil {
		return err
	}
	err = setKey(but, key, newBuf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

var txLocatorPrefix = []byte("txlocator-")

// GetTx returns a tx locator from a hash.
func (brt *BlockDBReadTransaction) GetTx(hash chainhash.Hash) (*primitives.TxLocator, error) {
	key := append(txLocatorPrefix, hash[:]...)
	locator := new(primitives.TxLocator)
	lbs, err := getKey(brt, key)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(lbs)
	err = locator.Decode(buf)
	if err != nil {
		return nil, err
	}
	return locator, nil
}

// SetTx stores a new locator for the specified hash.
func (but *BlockDBUpdateTransaction) SetTx(locator primitives.TxLocator) error {
	key := append(txLocatorPrefix, locator.TxHash[:]...)
	buf := bytes.NewBuffer([]byte{})
	err := locator.Encode(buf)
	if err != nil {
		return err
	}
	err = setKey(but, key, buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

var _ DB = &BlockDB{}
var _ DBUpdateTransaction = &BlockDBUpdateTransaction{}
var _ DBViewTransaction = &BlockDBReadTransaction{}

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
	GetFinalizedState() (*primitives.State, error)
	GetJustifiedState() (*primitives.State, error)
	GetBlockRow(chainhash.Hash) (*BlockNodeDisk, error)
	GetJustifiedHead() (chainhash.Hash, error)
	GetFinalizedHead() (chainhash.Hash, error)
	GetGenesisTime() (time.Time, error)
	GetAccountTxs([20]byte) (*primitives.AccountTxs, error)
	GetTx(chainhash.Hash) (*primitives.TxLocator, error)
}

// DBTransaction is a transaction to update the state of the database.
type DBUpdateTransaction interface {
	AddRawBlock(block *primitives.Block) error
	SetTip(chainhash.Hash) error
	SetFinalizedState(*primitives.State) error
	SetJustifiedState(*primitives.State) error
	SetBlockRow(*BlockNodeDisk) error
	SetJustifiedHead(chainhash.Hash) error
	SetFinalizedHead(chainhash.Hash) error
	SetGenesisTime(time.Time) error
	SetAccountTx([20]byte, chainhash.Hash) error
	SetTx(primitives.TxLocator) error
	DBViewTransaction
}
