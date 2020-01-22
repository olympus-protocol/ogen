package dbindex

import "github.com/dgraph-io/badger"

var (
	govIndexKeyPrefix = []byte("gov-")
)

type Gov struct {
	DB *badger.DB
}

func (i *Gov) Size() (int64, error) {
	var size int64
	err := i.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: govIndexKeyPrefix, PrefetchValues: false})
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

func (i *Gov) Remove(id []byte) error {
	err := i.DB.Update(func(txn *badger.Txn) error {
		key := append(govIndexKeyPrefix, id...)
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

func (i *Gov) Get() {}

func (i *Gov) GetAll() ([]byte, error) {
	var index []byte
	err := i.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: govIndexKeyPrefix})
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

func (i *Gov) Add() {}

func (i *Gov) Clear() {}
