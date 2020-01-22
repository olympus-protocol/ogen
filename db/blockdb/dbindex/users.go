package dbindex

import "github.com/dgraph-io/badger"

var (
	usersIndexKeyPrefix = []byte("user-")
)

type Users struct {
	DB *badger.DB
}

func (i *Users) Size() (int64, error) {
	var size int64
	err := i.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: usersIndexKeyPrefix, PrefetchValues: false})
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

func (i *Users) Remove(id []byte) error {
	err := i.DB.Update(func(txn *badger.Txn) error {
		key := append(usersIndexKeyPrefix, id...)
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

func (i *Users) Get() {}

func (i *Users) GetAll() ([]byte, error) {
	var index []byte
	err := i.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: usersIndexKeyPrefix})
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

func (i *Users) Add() {}

func (i *Users) Clear() {}
