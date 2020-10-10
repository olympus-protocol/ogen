package blockdb

import (
	"errors"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"path"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	blockBkt    = []byte("blocks")
	tipsBkt     = []byte("tip")
	statesBkt   = []byte("state")
	genesisBkt  = []byte("genesis")
	blockRowBkt = []byte("block_row")

	tipKey      = []byte("tip")
	finHeadKey  = []byte("finalized_head")
	jusHeadKey  = []byte("justified_head")
	finStateKey = []byte("finalized_state")
	jusStateKey = []byte("justified_state")
	genTimeKey  = []byte("genesis_key")
)

type boltDB struct {
	log       logger.Logger
	netParams *params.ChainParams

	lock     sync.Mutex
	db       *bolt.DB
	canClose sync.WaitGroup
	readTx   *bolt.Tx
}

// NewBoltDB returns a database instance with a rawBlockDatabase and boltDB to use on the selected path.
func NewBoltDB() (Database, error) {

	datapath := config.GlobalFlags.DataPath
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	db, err := bolt.Open(path.Join(datapath, "chain.db"), 0700, &bolt.Options{Timeout: 1 * time.Second, InitialMmapSize: 10e6})
	if err != nil {
		if err == bolt.ErrTimeout {
			return nil, errors.New("cannot obtain database lock, database may be in use by another process")
		}
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(blockBkt)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(tipsBkt)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(genesisBkt)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(statesBkt)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(blockRowBkt)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	readTx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	blockdb := &boltDB{
		log:       log,
		db:        db,
		netParams: netParams,
		readTx:    readTx,
	}

	return blockdb, nil
}

// Close closes the database.
func (db *boltDB) Close() {
	db.canClose.Wait()
	_ = db.db.Close()
}

// GetBlock gets a block from the database.
func (db *boltDB) GetBlock(hash chainhash.Hash) (*primitives.Block, error) {
	blockBytes, err := db.GetRawBlock(hash)
	if err != nil {
		return nil, err
	}
	block := new(primitives.Block)
	err = block.Unmarshal(blockBytes)
	return block, err
}

// GetRawBlock gets a block serialized from the database.
func (db *boltDB) GetRawBlock(hash chainhash.Hash) ([]byte, error) {
	blockBytes := db.readTx.Bucket(blockBkt).Get(hash[:])
	return blockBytes, nil
}

// AddRawBlock adds a raw block to the database.
func (db *boltDB) AddRawBlock(block *primitives.Block) error {
	db.canClose.Add(1)
	defer db.canClose.Done()
	blockHash := block.Hash()
	blockBytes, err := block.Marshal()
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(blockBkt).Put(blockHash[:], blockBytes)
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

// SetTip sets the current best tip of the blockchain.
func (db *boltDB) SetTip(c chainhash.Hash) error {
	db.canClose.Add(1)
	defer db.canClose.Done()
	err := db.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(tipsBkt).Put(tipKey, c[:])
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

// GetTip gets the current best tip of the blockchain.
func (db *boltDB) GetTip() (chainhash.Hash, error) {
	var tip chainhash.Hash
	tipBytes := db.readTx.Bucket(tipsBkt).Get(tipKey)
	copy(tip[:], tipBytes)
	return tip, nil
}

// SetJustifiedHead sets the latest justified head.
func (db *boltDB) SetJustifiedHead(c chainhash.Hash) error {
	db.canClose.Add(1)
	defer db.canClose.Done()
	err := db.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(tipsBkt).Put(jusHeadKey, c[:])
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

// GetJustifiedHead gets the latest justified head.
func (db *boltDB) GetJustifiedHead() (chainhash.Hash, error) {
	var head chainhash.Hash
	headBytes := db.readTx.Bucket(tipsBkt).Get(jusHeadKey)
	copy(head[:], headBytes)
	return head, nil
}

// SetFinalizedHead sets the finalized head of the blockchain.
func (db *boltDB) SetFinalizedHead(c chainhash.Hash) error {
	db.canClose.Add(1)
	defer db.canClose.Done()
	err := db.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(tipsBkt).Put(finHeadKey, c[:])
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

// GetFinalizedHead gets the finalized head of the blockchain.
func (db *boltDB) GetFinalizedHead() (chainhash.Hash, error) {
	var head chainhash.Hash
	headBytes := db.readTx.Bucket(tipsBkt).Get(finHeadKey)
	copy(head[:], headBytes)
	return head, nil
}

// SetFinalizedState sets the finalized state of the blockchain.
func (db *boltDB) SetFinalizedState(s state.State) error {
	db.canClose.Add(1)
	defer db.canClose.Done()
	buf, err := s.Marshal()
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(statesBkt).Put(finStateKey, buf)
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

// GetFinalizedState gets the finalized state of the blockchain.
func (db *boltDB) GetFinalizedState() (state.State, error) {
	stateBytes := db.readTx.Bucket(statesBkt).Get(finStateKey)
	s := state.NewEmptyState()
	err := s.Unmarshal(stateBytes)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// SetJustifiedState sets the justified state of the blockchain.
func (db *boltDB) SetJustifiedState(s state.State) error {
	db.canClose.Add(1)
	defer db.canClose.Done()
	buf, err := s.Marshal()
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(statesBkt).Put(jusStateKey, buf)
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

// GetJustifiedState gets the justified state of the blockchain.
func (db *boltDB) GetJustifiedState() (state.State, error) {
	stateBytes := db.readTx.Bucket(statesBkt).Get(jusStateKey)
	s := state.NewEmptyState()
	err := s.Unmarshal(stateBytes)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// SetBlockRow sets a block row on disk to store the block index.
func (db *boltDB) SetBlockRow(disk *primitives.BlockNodeDisk) error {
	db.canClose.Add(1)
	defer db.canClose.Done()
	diskSer, err := disk.Marshal()
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(blockRowBkt).Put(disk.Hash[:], diskSer)
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

// GetBlockRow gets the block row on disk.
func (db *boltDB) GetBlockRow(c chainhash.Hash) (*primitives.BlockNodeDisk, error) {
	diskSer := db.readTx.Bucket(blockRowBkt).Get(c[:])
	d := new(primitives.BlockNodeDisk)
	err := d.Unmarshal(diskSer)
	return d, err
}

// SetGenesisTime sets the genesis time of the blockchain.
func (db *boltDB) SetGenesisTime(t time.Time) error {
	db.canClose.Add(1)
	defer db.canClose.Done()
	bs, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(genesisBkt).Put(genTimeKey, bs)
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

// GetGenesisTime gets the genesis time of the blockchain.
func (db *boltDB) GetGenesisTime() (time.Time, error) {
	genBytes := db.readTx.Bucket(genesisBkt).Get(genTimeKey)
	var t time.Time
	err := t.UnmarshalBinary(genBytes)
	return t, err
}
