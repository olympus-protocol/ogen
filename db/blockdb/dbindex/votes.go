package dbindex

import "github.com/dgraph-io/badger"

var (
	voteIndexKeyPrefix = []byte("vote-")
)

type Votes struct {
	DB *badger.DB
}

func (i *Votes) Size() (int64, error) {
	var size int64
	err := i.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: voteIndexKeyPrefix, PrefetchValues: false})
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			size += it.Item().ValueSize()
		}
		return nil
	})
	if err != nil {
		return 0, nil
	}
	return size, nil
}

func (i *Votes) Remove(id []byte) error {
	err := i.DB.Update(func(txn *badger.Txn) error {
		key := append(voteIndexKeyPrefix, id...)
		err := txn.Delete(key)
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

func (i *Votes) Get() {}

func (i *Votes) GetAll() ([]byte, error) {
	var index []byte
	err := i.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: voteIndexKeyPrefix})
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			data, err := it.Item().ValueCopy(nil)
			if err != nil {
				return err
			}
			index = append(index, data...)
		}
		it.Close()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return index, nil
}

func (i *Votes) Add() {}

func (i *Votes) Clear() {}
