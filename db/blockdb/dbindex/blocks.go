package dbindex

import (
	"bytes"
	"github.com/dgraph-io/badger"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/utils/chainhash"
)

var (
	blockIndexKeyPrefix = []byte("block-")
)

type Blocks struct {
	DB *badger.DB
}

func (i *Blocks) Size() (int64, error) {
	var size int64
	err := i.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: blockIndexKeyPrefix, PrefetchValues: false})
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

func (i *Blocks) Add(locator []byte, header *p2p.BlockHeader) error {
	err := i.DB.Update(func(txn *badger.Txn) error {
		key := append(blockIndexKeyPrefix, header.PrevBlockHash.CloneBytes()...)
		buf := bytes.NewBuffer([]byte{})
		buf.Write(locator)
		err := header.Serialize(buf)
		if err != nil {
			return err
		}
		entry := badger.NewEntry(key, buf.Bytes())
		err = txn.SetEntry(entry)
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

func (i *Blocks) Get(prevBlockHash chainhash.Hash) ([]byte, error) {
	var blockRow []byte
	err := i.DB.View(func(txn *badger.Txn) error {
		key := append(blockIndexKeyPrefix, prevBlockHash.CloneBytes()...)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		blockRow, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blockRow, nil
}

func (i *Blocks) Clear() {}
