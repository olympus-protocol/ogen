package blockdb

import (
	"errors"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"sync"
	"time"
)

var (
	ErrorNoBlock    = errors.New("no block data")
	ErrorNoBlockRow = errors.New("no block row data")
)

type memoryDB struct {
	genesisTime time.Time

	justifiedState state.State
	justifiedHead  chainhash.Hash
	finalizedState state.State
	finalizedHead  chainhash.Hash

	tip chainhash.Hash

	blockRows     map[chainhash.Hash]*primitives.BlockNodeDisk
	blockRowsLock sync.RWMutex

	blocks     map[chainhash.Hash]*primitives.Block
	blocksLock sync.RWMutex

	log    logger.Logger
	params params.ChainParams
	lock   sync.RWMutex
}

// NewBadgerDB returns a database instance with a rawBlockDatabase and BadgerDB to use on the selected path.
func NewMemoryDB(params params.ChainParams, log logger.Logger) (Database, error) {
	memoryDB := &memoryDB{
		log:       log,
		params:    params,
		blockRows: make(map[chainhash.Hash]*primitives.BlockNodeDisk),
		blocks:    make(map[chainhash.Hash]*primitives.Block),
	}
	return memoryDB, nil
}

// Close closes the database.
func (db *memoryDB) Close() {}

// GetBlock gets a block from the database.
func (db *memoryDB) GetBlock(hash chainhash.Hash) (*primitives.Block, error) {
	db.blocksLock.Lock()
	defer db.blocksLock.Unlock()
	block, ok := db.blocks[hash]
	if !ok {
		return nil, ErrorNoBlock
	}
	return block, nil
}

// GetRawBlock gets a block serialized from the database.
func (db *memoryDB) GetRawBlock(hash chainhash.Hash) ([]byte, error) {
	db.blocksLock.Lock()
	defer db.blocksLock.Unlock()
	block, ok := db.blocks[hash]
	if !ok {
		return nil, ErrorNoBlock
	}
	blockRaw, err := block.Marshal()
	if err != nil {
		return nil, err
	}
	return blockRaw, nil
}

// AddRawBlock adds a raw block to the database.
func (db *memoryDB) AddRawBlock(block *primitives.Block) error {
	blockHash := block.Hash()
	db.blocksLock.Lock()
	defer db.blocksLock.Unlock()
	db.blocks[blockHash] = block
	return nil
}

// SetTip sets the current best tip of the blockchain.
func (db *memoryDB) SetTip(c chainhash.Hash) error {
	db.tip = c
	return nil
}

// GetTip gets the current best tip of the blockchain.
func (db *memoryDB) GetTip() (chainhash.Hash, error) {
	return db.tip, nil
}

// SetFinalizedState sets the finalized state of the blockchain.
func (db *memoryDB) SetFinalizedState(s state.State) error {
	db.finalizedState = s
	return nil
}

// GetFinalizedState gets the finalized state of the blockchain.
func (db *memoryDB) GetFinalizedState() (state.State, error) {
	return db.finalizedState, nil
}

// SetJustifiedState sets the justified state of the blockchain.
func (db *memoryDB) SetJustifiedState(s state.State) error {
	db.justifiedState = s
	return nil
}

// GetJustifiedState gets the justified state of the blockchain.
func (db *memoryDB) GetJustifiedState() (state.State, error) {
	return db.justifiedState, nil
}

// SetBlockRow sets a block row on disk to store the block index.
func (db *memoryDB) SetBlockRow(disk *primitives.BlockNodeDisk) error {
	db.blockRowsLock.Lock()
	defer db.blockRowsLock.Unlock()

	db.blockRows[disk.Hash] = disk
	return nil
}

// GetBlockRow gets the block row on disk.
func (db *memoryDB) GetBlockRow(c chainhash.Hash) (*primitives.BlockNodeDisk, error) {
	db.blockRowsLock.Lock()
	defer db.blockRowsLock.Unlock()
	row, ok := db.blockRows[c]
	if !ok {
		return nil, ErrorNoBlockRow
	}
	return row, nil
}

// SetJustifiedHead sets the latest justified head.
func (db *memoryDB) SetJustifiedHead(c chainhash.Hash) error {
	db.justifiedHead = c
	return nil
}

// GetJustifiedHead gets the latest justified head.
func (db *memoryDB) GetJustifiedHead() (chainhash.Hash, error) {
	return db.justifiedHead, nil
}

// SetFinalizedHead sets the finalized head of the blockchain.
func (db *memoryDB) SetFinalizedHead(c chainhash.Hash) error {
	db.finalizedHead = c
	return nil
}

// GetFinalizedHead gets the finalized head of the blockchain.
func (db *memoryDB) GetFinalizedHead() (chainhash.Hash, error) {
	return db.finalizedHead, nil
}

// SetGenesisTime sets the genesis time of the blockchain.
func (db *memoryDB) SetGenesisTime(t time.Time) error {
	db.genesisTime = t
	return nil
}

// GetGenesisTime gets the genesis time of the blockchain.
func (db *memoryDB) GetGenesisTime() (time.Time, error) {
	return db.genesisTime, nil
}
