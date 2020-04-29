package blockdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"

	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type BlockDB struct {
	log      *logger.Logger
	badgerdb *badger.DB
	params   params.ChainParams

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
	dbOptions := badger.DefaultOptions(path + "/db")
	dbOptions.Logger = nil
	badgerdb, err := badger.Open(dbOptions)
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

// GetRawBlock gets a block from the database.
func (brt *BlockDBReadTransaction) GetRawBlock(hash chainhash.Hash) (*primitives.Block, error) {
	blockBytes, err := getKey(brt, hash[:])
	if err != nil {
		return nil, err
	}

	block := new(primitives.Block)
	err = block.Decode(bytes.NewBuffer(blockBytes))
	return block, err
}

// AddRawBlock adds a raw block to the database.
func (but *BlockDBUpdateTransaction) AddRawBlock(block *primitives.Block) error {
	blockHash := block.Hash()
	blockBytes := bytes.NewBuffer([]byte{})
	err := block.Encode(blockBytes)
	if err != nil {
		return err
	}
	return setKey(but, blockHash[:], blockBytes.Bytes())
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

// GetLatestVote gets the latest vote by a validator.
func (brt *BlockDBReadTransaction) GetLatestVote(validator uint32) (*primitives.MultiValidatorVote, error) {
	var validatorBytes [4]byte
	binary.BigEndian.PutUint32(validatorBytes[:], validator)
	key := append(latestVotePrefix, validatorBytes[:]...)

	voteSer, err := getKey(brt, key)
	if err != nil {
		return nil, err
	}

	vote := new(primitives.MultiValidatorVote)
	err = vote.Deserialize(bytes.NewBuffer(voteSer))
	return vote, err
}

// SetLatestVoteIfNeeded sets the latest for a validator.
func (but *BlockDBUpdateTransaction) SetLatestVoteIfNeeded(validators []uint32, vote *primitives.MultiValidatorVote) error {
	buf := bytes.NewBuffer([]byte{})

	err := vote.Serialize(buf)
	if err != nil {
		return err
	}
	for _, validator := range validators {
		var validatorBytes [4]byte
		binary.BigEndian.PutUint32(validatorBytes[:], validator)
		key := append(latestVotePrefix, validatorBytes[:]...)

		existingItem, err := but.transaction.Get(key)
		if err == badger.ErrKeyNotFound {
			err := but.transaction.Set(key, buf.Bytes())
			if err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}

		existingBytes, err := existingItem.ValueCopy(nil)
		if err != nil {
			return err
		}
		existingBytesBuf := bytes.NewBuffer(existingBytes)

		oldVote := new(primitives.MultiValidatorVote)
		err = oldVote.Deserialize(existingBytesBuf)
		if err != nil {
			return err
		}

		if oldVote.Data.Slot >= vote.Data.Slot {
			continue
		}

		if err := but.transaction.Set(key, buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

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
	diskSer, err := disk.Serialize()
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
	err = d.Deserialize(diskSer)
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
	GetRawBlock(hash chainhash.Hash) (*primitives.Block, error)
	GetLatestVote(validator uint32) (*primitives.MultiValidatorVote, error)
	GetTip() (chainhash.Hash, error)
	GetFinalizedState() (*primitives.State, error)
	GetJustifiedState() (*primitives.State, error)
	GetBlockRow(chainhash.Hash) (*BlockNodeDisk, error)
	GetJustifiedHead() (chainhash.Hash, error)
	GetFinalizedHead() (chainhash.Hash, error)
	GetGenesisTime() (time.Time, error)
}

// DBTransaction is a transaction to update the state of the database.
type DBUpdateTransaction interface {
	AddRawBlock(block *primitives.Block) error
	SetLatestVoteIfNeeded(validators []uint32, vote *primitives.MultiValidatorVote) error
	SetTip(chainhash.Hash) error
	SetFinalizedState(*primitives.State) error
	SetJustifiedState(*primitives.State) error
	SetBlockRow(*BlockNodeDisk) error
	SetJustifiedHead(chainhash.Hash) error
	SetFinalizedHead(chainhash.Hash) error
	SetGenesisTime(time.Time) error
	DBViewTransaction
}
