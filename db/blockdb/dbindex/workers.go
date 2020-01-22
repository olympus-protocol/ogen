package dbindex

import "github.com/dgraph-io/badger"

var (
	workerIndexKeyPrefix = []byte("worker-")
)

type Workers struct {
	DB *badger.DB
}

func (i *Workers) Size() (int64, error) {
	var size int64
	err := i.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: workerIndexKeyPrefix, PrefetchValues: false})
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			size += it.Item().ValueSize()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (i *Workers) Remove(id []byte) error {
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

func (i *Workers) Get() {}

func (i *Workers) GetAll() ([]byte, error) {
	var index []byte
	err := i.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: workerIndexKeyPrefix})
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

func (i *Workers) Add() {}

func (i *Workers) Clear() {}
