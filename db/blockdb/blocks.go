package blockdb

import (
	"bytes"
	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var (
	blockIndexKeyPrefix = []byte("block-")
)

func (bdb *BlockDB) addBlock(locator []byte, header primitives.BlockHeader) error {
	err := bdb.badgerdb.Update(func(txn *badger.Txn) error {
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

func (bdb *BlockDB) getBlock(prevBlockHash chainhash.Hash) ([]byte, error) {
	var blockRow []byte
	err := bdb.badgerdb.View(func(txn *badger.Txn) error {
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

func (i *BlockDB) Clear() {}
