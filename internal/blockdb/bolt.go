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

	tipKey = []byte("tip")
	finHeadKey = []byte("finalized_head")
	jusHeadKey = []byte("justified_head")
	finStateKey = []byte("finalized_state")
	jusStateKey = []byte("justified_state")
	genTimeKey = []byte("genesis_key")
)

type boltDB struct {
	log       logger.Logger
	netParams *params.ChainParams

	lock     sync.Mutex
	db       *bolt.DB
	canClose sync.WaitGroup
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
	db.AllocSize = 8 * 1024 * 1024

	blockdb := &boltDB{
		log:       log,
		db:        db,
		netParams: netParams,
	}

	err = blockdb.createBuckets()
	if err != nil {
		return nil, err
	}
	return blockdb, nil
}

func (db *boltDB) createBuckets() error {
	return db.db.Update(func(tx *bolt.Tx) error {
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
}

// Close closes the database.
func (db *boltDB) Close() {
	db.canClose.Wait()
	_ = db.db.Close()
}

// GetBlock gets a block from the database.
func (db *boltDB) GetBlock(hash chainhash.Hash) (*primitives.Block, error) {
	var blockBytes []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blockBkt)
		blockBytes = bkt.Get(hash[:])
		return nil
	})
	if err != nil {
		return nil, err
	}
	block := new(primitives.Block)
	err = block.Unmarshal(blockBytes)
	return block, err
}

// GetRawBlock gets a block serialized from the database.
func (db *boltDB) GetRawBlock(hash chainhash.Hash) ([]byte, error) {
	var blockBytes []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blockBkt)
		blockBytes = bkt.Get(hash[:])
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blockBytes, err
}

// AddRawBlock adds a raw block to the database.
func (db *boltDB) AddRawBlock(block *primitives.Block) error {
	blockHash := block.Hash()
	blockBytes, err := block.Marshal()
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blockBkt)
		err = bkt.Put(blockHash[:], blockBytes)
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
	err := db.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(tipsBkt)
		err := bkt.Put(tipKey, c[:])
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
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(tipsBkt)
		tipBytes := bkt.Get(tipKey)
		copy(tip[:], tipBytes)
		return nil
	})
	if err != nil {
		return chainhash.Hash{}, err
	}
	return tip, nil
}

// SetFinalizedState sets the finalized state of the blockchain.
func (db *boltDB) SetFinalizedState(s state.State) error {
	buf, err := s.Marshal()
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(statesBkt)
		err = bkt.Put(finStateKey, buf)
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
	var stateBytes []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(statesBkt)
		stateBytes = bkt.Get(finStateKey)
		return nil
	})
	if err != nil {
		return nil, err
	}
	s := state.NewEmptyState()
	err = s.Unmarshal(stateBytes)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// SetJustifiedState sets the justified state of the blockchain.
func (db *boltDB) SetJustifiedState(s state.State) error {
	buf, err := s.Marshal()
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(statesBkt)
		err = bkt.Put(jusStateKey, buf)
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
	var stateBytes []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(statesBkt)
		stateBytes = bkt.Get(jusStateKey)
		return nil
	})
	if err != nil {
		return nil, err
	}
	s := state.NewEmptyState()
	err = s.Unmarshal(stateBytes)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// SetBlockRow sets a block row on disk to store the block index.
func (db *boltDB) SetBlockRow(disk *primitives.BlockNodeDisk) error {
	diskSer, err := disk.Marshal()
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blockRowBkt)
		err = bkt.Put(disk.Hash[:], diskSer)
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
	var diskSer []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blockRowBkt)
		diskSer = bkt.Get(c[:])
		return nil
	})
	if err != nil {
		return nil, err
	}

	d := new(primitives.BlockNodeDisk)
	err = d.Unmarshal(diskSer)
	return d, err
}

// SetJustifiedHead sets the latest justified head.
func (db *boltDB) SetJustifiedHead(c chainhash.Hash) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(tipsBkt)
		err := bkt.Put(jusHeadKey, c[:])
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
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(tipsBkt)
		headBytes := bkt.Get(jusHeadKey)
		copy(head[:], headBytes)
		return nil
	})
	if err != nil {
		return chainhash.Hash{}, err
	}
	return head, nil
}

// SetFinalizedHead sets the finalized head of the blockchain.
func (db *boltDB) SetFinalizedHead(c chainhash.Hash) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(tipsBkt)
		err := bkt.Put(finHeadKey, c[:])
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
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(tipsBkt)
		headBytes := bkt.Get(finHeadKey)
		copy(head[:], headBytes)
		return nil
	})
	if err != nil {
		return chainhash.Hash{}, err
	}
	return head, nil
}

// SetGenesisTime sets the genesis time of the blockchain.
func (db *boltDB) SetGenesisTime(t time.Time) error {
	bs, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(genesisBkt)
		err := bkt.Put(genTimeKey, bs)
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
	var genBytes []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(genesisBkt)
		genBytes = bkt.Get(genTimeKey)
		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	var t time.Time
	err = t.UnmarshalBinary(genBytes)
	return t, err
}
