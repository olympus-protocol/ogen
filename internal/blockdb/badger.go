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

	"github.com/dgraph-io/badger"
)

type badgerDB struct {
	log       logger.Logger
	netParams *params.ChainParams

	lock     sync.Mutex
	badgerdb *badger.DB
	canClose sync.WaitGroup
}

// NewBadgerDB returns a database instance with a rawBlockDatabase and BadgerDB to use on the selected path.
func NewBadgerDB() (Database, error) {
	datapath := config.GlobalFlags.DataPath
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	badgerdb, err := badger.Open(badger.DefaultOptions(datapath + "/chain").WithLogger(nil))
	if err != nil {
		return nil, err
	}
	blockdb := &badgerDB{
		log:       log,
		badgerdb:  badgerdb,
		netParams: netParams,
	}
	return blockdb, nil
}

// Close closes the database.
func (db *badgerDB) Close() {
	db.canClose.Wait()
	_ = db.badgerdb.Close()
}

// GetBlock gets a block from the database.
func (db *badgerDB) GetBlock(hash chainhash.Hash) (*primitives.Block, error) {
	blockBytes, err := db.getKey(hash[:])
	if err != nil {
		return nil, err
	}

	block := new(primitives.Block)
	err = block.Unmarshal(blockBytes)
	return block, err
}

// GetRawBlock gets a block serialized from the database.
func (db *badgerDB) GetRawBlock(hash chainhash.Hash) ([]byte, error) {
	blockBytes, err := db.getKey(hash[:])
	if err != nil {
		return nil, err
	}
	return blockBytes, err
}

// AddRawBlock adds a raw block to the database.
func (db *badgerDB) AddRawBlock(block *primitives.Block, isCheck bool) error {
	blockHash := block.Hash()
	blockBytes, err := block.Marshal()
	if err != nil {
		return err
	}
	if isCheck {
		return nil
	}
	return db.setKey(blockHash[:], blockBytes)
}

var tipKey = []byte("chain-tip")

// SetTip sets the current best tip of the blockchain.
func (db *badgerDB) SetTip(c chainhash.Hash) error {
	return db.setKeyHash(tipKey, c)
}

// GetTip gets the current best tip of the blockchain.
func (db *badgerDB) GetTip() (chainhash.Hash, error) {
	return db.getKeyHash(tipKey)
}

var finalizedStateKey = []byte("finalized-state")

// SetFinalizedState sets the finalized state of the blockchain.
func (db *badgerDB) SetFinalizedState(s state.State) (state.State, error) {
	buf, err := s.Marshal()
	if err != nil {
		return nil, err
	}
	oldState, _ := db.GetFinalizedState()
	return oldState, db.setKey(finalizedStateKey, buf)
}

// GetFinalizedState gets the finalized state of the blockchain.
func (db *badgerDB) GetFinalizedState() (state.State, error) {
	stateBytes, err := db.getKey(finalizedStateKey)
	if err != nil {
		return nil, err
	}
	s := state.NewEmptyState()
	err = s.Unmarshal(stateBytes)
	return s, err
}

var justifiedStateKey = []byte("justified-state")

// SetJustifiedState sets the justified state of the blockchain.
func (db *badgerDB) SetJustifiedState(s state.State) (state.State, error) {
	buf, err := s.Marshal()
	if err != nil {
		return nil, err
	}
	oldState, _ := db.GetJustifiedState()
	return oldState, db.setKey(justifiedStateKey, buf)
}

// GetJustifiedState gets the justified state of the blockchain.
func (db *badgerDB) GetJustifiedState() (state.State, error) {
	stateBytes, err := db.getKey(justifiedStateKey)
	if err != nil {
		return nil, err
	}
	s := state.NewEmptyState()
	err = s.Unmarshal(stateBytes)
	return s, err
}

var blockRowPrefix = []byte("block-row")

// SetBlockRow sets a block row on disk to store the block index.
func (db *badgerDB) SetBlockRow(disk *primitives.BlockNodeDisk) error {
	key := append(blockRowPrefix, disk.Hash[:]...)
	diskSer, err := disk.Marshal()
	if err != nil {
		return err
	}
	return db.setKey(key, diskSer)
}

// GetBlockRow gets the block row on disk.
func (db *badgerDB) GetBlockRow(c chainhash.Hash) (*primitives.BlockNodeDisk, error) {
	key := append(blockRowPrefix, c[:]...)
	diskSer, err := db.getKey(key)
	if err != nil {
		return nil, err
	}

	d := new(primitives.BlockNodeDisk)
	err = d.Unmarshal(diskSer)
	return d, err
}

var justifiedHeadKey = []byte("justified-head")

// SetJustifiedHead sets the latest justified head.
func (db *badgerDB) SetJustifiedHead(c chainhash.Hash) (chainhash.Hash, error) {
	oldHead, _ := db.GetJustifiedHead()
	return oldHead, db.setKeyHash(justifiedHeadKey, c)
}

// GetJustifiedHead gets the latest justified head.
func (db *badgerDB) GetJustifiedHead() (chainhash.Hash, error) {
	return db.getKeyHash(justifiedHeadKey)
}

var finalizedHeadKey = []byte("finalized-head")

// SetFinalizedHead sets the finalized head of the blockchain.
func (db *badgerDB) SetFinalizedHead(c chainhash.Hash) (chainhash.Hash, error) {
	oldFinalized, _ := db.getKeyHash(finalizedHeadKey)
	return oldFinalized, db.setKeyHash(finalizedHeadKey, c)
}

// GetFinalizedHead gets the finalized head of the blockchain.
func (db *badgerDB) GetFinalizedHead() (chainhash.Hash, error) {
	return db.getKeyHash(finalizedHeadKey)
}

var genesisTimeKey = []byte("genesisTime")

// SetGenesisTime sets the genesis time of the blockchain.
func (db *badgerDB) SetGenesisTime(t time.Time) error {
	bs, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	return db.setKey(genesisTimeKey, bs)
}

// GetGenesisTime gets the genesis time of the blockchain.
func (db *badgerDB) GetGenesisTime() (time.Time, error) {
	bs, err := db.getKey(genesisTimeKey)
	if err != nil {
		return time.Time{}, err
	}

	var t time.Time
	err = t.UnmarshalBinary(bs)
	return t, err
}

func (db *badgerDB) getKeyHash(key []byte) (chainhash.Hash, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.canClose.Add(1)
	defer db.canClose.Done()
	var out chainhash.Hash
	err := db.badgerdb.View(func(txn *badger.Txn) error {
		i, err := txn.Get(key)
		if err != nil {
			return err
		}
		h, err := i.ValueCopy(nil)
		if err != nil {
			return err
		}
		copy(out[:], h)
		return nil
	})
	if err != nil {
		return chainhash.Hash{}, err
	}
	return out, nil
}

func (db *badgerDB) getKey(key []byte) ([]byte, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.canClose.Add(1)
	defer db.canClose.Done()
	var out []byte
	err := db.badgerdb.View(func(txn *badger.Txn) error {
		i, err := txn.Get(key)
		if err != nil {
			return err
		}
		out, err = i.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (db *badgerDB) setKeyHash(key []byte, to chainhash.Hash) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.canClose.Add(1)
	defer db.canClose.Done()
	err := db.badgerdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, to[:])
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *badgerDB) setKey(key []byte, to []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.canClose.Add(1)
	defer db.canClose.Done()
	err := db.badgerdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, to)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
