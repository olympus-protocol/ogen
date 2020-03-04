package blockdb

import (
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

type DB interface {
	Close()
	GetRawBlock(locator BlockLocation, hash chainhash.Hash) ([]byte, error)
	AddRawBlock(block *primitives.Block) (*BlockLocation, error)
	Clear()
}
