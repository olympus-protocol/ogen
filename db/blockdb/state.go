package blockdb

import "github.com/dgraph-io/badger"

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
