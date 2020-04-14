package blockdb

import (
	"bytes"
	"encoding/binary"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var chainStateKey = []byte("chain-state-")

func (bdb *BlockDB) GetStateSnap() ([]byte, error) {
	var chainState []byte
	err := bdb.badgerdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(chainStateKey)
		if err != nil {
			return err
		}
		chainState, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return chainState, nil
}

func (bdb *BlockDB) SetStateSnap(data []byte) error {
	err := bdb.badgerdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(chainStateKey, data)
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

// BlockNodeDisk is a block node stored on disk.
type BlockNodeDisk struct {
	Header    primitives.BlockHeader
	Locator   BlockLocation
	StateRoot chainhash.Hash
	Height    int32
	Hash      chainhash.Hash
	Parent    chainhash.Hash
}

func (bnd *BlockNodeDisk) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	err := bnd.Header.Serialize(buf)
	if err != nil {
		return nil, err
	}

	err = bnd.Locator.Serialize(buf)
	if err != nil {
		return nil, err
	}

	err = serializer.WriteElements(buf, bnd.StateRoot, bnd.Height, bnd.Hash, bnd.Parent)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (bnd *BlockNodeDisk) Deserialize(b []byte) error {
	buf := bytes.NewBuffer(b)

	err := bnd.Header.Deserialize(buf)
	if err != nil {
		return err
	}

	err = bnd.Locator.Deserialize(buf)
	if err != nil {
		return err
	}

	err = serializer.ReadElements(buf, &bnd.StateRoot, &bnd.Height, &bnd.Hash, &bnd.Parent)
	if err != nil {
		return err
	}

	return nil
}

func getKeyHash(db *badger.DB, key []byte) (chainhash.Hash, error) {
	var out chainhash.Hash
	err := db.View(func(txn *badger.Txn) error {
		i, err := txn.Get(key)
		if err != nil {
			return err
		}
		_, err = i.ValueCopy(out[:])
		return err
	})
	if err != nil {
		return out, err
	}
	return out, nil
}

func getKey(db *badger.DB, key []byte) ([]byte, error) {
	var out []byte
	err := db.View(func(txn *badger.Txn) error {
		i, err := txn.Get(key)
		if err != nil {
			return err
		}
		out, err = i.ValueCopy(out)
		return err
	})
	if err != nil {
		return out, err
	}
	return out, nil
}

func setKeyHash(db *badger.DB, key []byte, to chainhash.Hash) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, to[:])
	})
}

func setKey(db *badger.DB, key []byte, to []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, to)
	})
}

var latestVotePrefix = []byte("latest-vote")

func (bdb *BlockDB) GetLatestVote(validator uint32) (*primitives.MultiValidatorVote, error) {
	var validatorBytes [4]byte
	binary.BigEndian.PutUint32(validatorBytes[:], validator)
	key := append(latestVotePrefix, validatorBytes[:]...)

	voteSer, err := getKey(bdb.badgerdb, key)
	if err != nil {
		return nil, err
	}

	vote := new(primitives.MultiValidatorVote)
	err = vote.Deserialize(bytes.NewBuffer(voteSer))
	return vote, err
}

func (bdb *BlockDB) SetLatestVoteIfNeeded(validators []uint32, vote *primitives.MultiValidatorVote) error {
	buf := bytes.NewBuffer([]byte{})

	err := vote.Serialize(buf)
	if err != nil {
		return err
	}
	return bdb.badgerdb.Update(func(txn *badger.Txn) error {
		for _, validator := range validators {
			var validatorBytes [4]byte
			binary.BigEndian.PutUint32(validatorBytes[:], validator)
			key := append(latestVotePrefix, validatorBytes[:]...)

			existingItem, err := txn.Get(key)
			if err == badger.ErrKeyNotFound {
				err := txn.Set(key, buf.Bytes())
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

			if err := txn.Set(key, buf.Bytes()); err != nil {
				return err
			}
		}

		return nil
	})
}

var tipKey = []byte("chain-tip")

func (bdb *BlockDB) SetTip(c chainhash.Hash) error {
	return setKeyHash(bdb.badgerdb, tipKey, c)
}

func (bdb *BlockDB) GetTip() (chainhash.Hash, error) {
	return getKeyHash(bdb.badgerdb, tipKey)
}

var finalizedStateKey = []byte("finalized-state")

func (bdb *BlockDB) SetFinalizedState(s *primitives.State) error {
	buf := bytes.NewBuffer([]byte{})
	if err := s.Serialize(buf); err != nil {
		return err
	}

	return setKey(bdb.badgerdb, finalizedStateKey, buf.Bytes())
}

func (bdb *BlockDB) GetFinalizedState() (*primitives.State, error) {
	stateBytes, err := getKey(bdb.badgerdb, finalizedStateKey)
	if err != nil {
		return nil, err
	}
	stateBuf := bytes.NewBuffer(stateBytes)
	state := new(primitives.State)
	err = state.Deserialize(stateBuf)
	return state, err
}

var justifiedStateKey = []byte("justified-state")

func (bdb *BlockDB) SetJustifiedState(s *primitives.State) error {
	buf := bytes.NewBuffer([]byte{})
	if err := s.Serialize(buf); err != nil {
		return err
	}

	return setKey(bdb.badgerdb, justifiedStateKey, buf.Bytes())
}

func (bdb *BlockDB) GetJustifiedState() (*primitives.State, error) {
	stateBytes, err := getKey(bdb.badgerdb, justifiedStateKey)
	if err != nil {
		return nil, err
	}
	stateBuf := bytes.NewBuffer(stateBytes)
	state := new(primitives.State)
	err = state.Deserialize(stateBuf)
	return state, err
}

var blockRowPrefix = []byte("block-row")

func (bdb *BlockDB) SetBlockRow(disk *BlockNodeDisk) error {
	key := append(blockRowPrefix, disk.Hash[:]...)
	diskSer, err := disk.Serialize()
	if err != nil {
		return err
	}
	return setKey(bdb.badgerdb, key, diskSer)
}

func (bdb *BlockDB) GetBlockRow(c chainhash.Hash) (*BlockNodeDisk, error) {
	key := append(blockRowPrefix, c[:]...)
	diskSer, err := getKey(bdb.badgerdb, key)
	if err != nil {
		return nil, err
	}

	d := new(BlockNodeDisk)
	err = d.Deserialize(diskSer)
	return d, err
}

var justifiedHeadKey = []byte("justified-head")

func (bdb *BlockDB) SetJustifiedHead(c chainhash.Hash) error {
	return setKeyHash(bdb.badgerdb, justifiedHeadKey, c)
}

func (bdb *BlockDB) GetJustifiedHead() (chainhash.Hash, error) {
	return getKeyHash(bdb.badgerdb, justifiedHeadKey)
}

var finalizedHeadKey = []byte("finalized-head")

func (bdb *BlockDB) SetFinalizedHead(c chainhash.Hash) error {
	return setKeyHash(bdb.badgerdb, finalizedHeadKey, c)
}

func (bdb *BlockDB) GetFinalizedHead() (chainhash.Hash, error) {
	return getKeyHash(bdb.badgerdb, finalizedHeadKey)
}

var genesisTimeKey = []byte("genesisTime")

func (bdb *BlockDB) SetGenesisTime(t time.Time) error {
	bs, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	return setKey(bdb.badgerdb, genesisTimeKey, bs)
}

func (bdb *BlockDB) GetGenesisTime() (time.Time, error) {
	bs, err := getKey(bdb.badgerdb, genesisTimeKey)
	if err != nil {
		return time.Time{}, err
	}

	var t time.Time
	err = t.UnmarshalBinary(bs)
	return t, err
}

var _ DB = &BlockDB{}

// DB is the interface to store various elements of the state of the chain.
type DB interface {
	Close()
	GetRawBlock(locator BlockLocation, hash chainhash.Hash) ([]byte, error)
	AddRawBlock(block *primitives.Block) (*BlockLocation, error)
	GetLatestVote(validator uint32) (*primitives.MultiValidatorVote, error)
	SetLatestVoteIfNeeded(validators []uint32, vote *primitives.MultiValidatorVote) error
	SetTip(chainhash.Hash) error
	GetTip() (chainhash.Hash, error)
	SetFinalizedState(*primitives.State) error
	GetFinalizedState() (*primitives.State, error)
	SetJustifiedState(*primitives.State) error
	GetJustifiedState() (*primitives.State, error)
	SetBlockRow(*BlockNodeDisk) error
	GetBlockRow(chainhash.Hash) (*BlockNodeDisk, error)
	SetJustifiedHead(chainhash.Hash) error
	GetJustifiedHead() (chainhash.Hash, error)
	SetFinalizedHead(chainhash.Hash) error
	GetFinalizedHead() (chainhash.Hash, error)
	SetGenesisTime(time.Time) error
	GetGenesisTime() (time.Time, error)
	Clear()
}
