package csmt_test

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/internal/csmt"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomWritesRollbackCommitBadger(t *testing.T) {
	badgerdb, err := badger.Open(badger.DefaultOptions("./badger-test"))
	assert.NoError(t, err)

	err = badgerdb.DropAll()
	assert.NoError(t, err)

	defer badgerdb.Close()

	under := csmt.NewBadgerTreeDB(badgerdb)

	underlyingTree := csmt.NewTree(under)

	var treeRoot chainhash.Hash

	err = underlyingTree.Update(func(tx csmt.TreeTransactionAccess) error {
		for i := 0; i < 200; i++ {
			err := tx.Set(ch(fmt.Sprintf("key%d", i)), ch(fmt.Sprintf("val%d", i)))
			if err != nil {
				return err
			}
		}

		initialRoot, err := tx.Hash()
		if err != nil {
			return err
		}

		treeRoot = *initialRoot

		return nil
	})

	assert.NoError(t, err)

	for i := 0; i < 100; i++ {
		cachedTreeDB, err := csmt.NewTreeMemoryCache(under)
		assert.NoError(t, err)

		cachedTree := csmt.NewTree(cachedTreeDB)

		err = cachedTree.Update(func(tx csmt.TreeTransactionAccess) error {
			for newVal := 198; newVal < 202; newVal++ {
				err := tx.Set(ch(fmt.Sprintf("key%d", i)), ch(fmt.Sprintf("val2%d", newVal)))
				assert.NoError(t, err)

			}

			return nil
		})
		assert.NoError(t, err)

	}

	var underlyingHash chainhash.Hash

	err = underlyingTree.View(func(tx csmt.TreeTransactionAccess) error {
		h, err := tx.Hash()
		if err != nil {
			return err
		}
		underlyingHash = *h

		return nil
	})
	assert.NoError(t, err)

	assert.Equal(t, underlyingHash, treeRoot)

	cachedTreeDB, err := csmt.NewTreeMemoryCache(under)
	assert.NoError(t, err)

	cachedTree := csmt.NewTree(cachedTreeDB)

	err = cachedTree.Update(func(tx csmt.TreeTransactionAccess) error {
		for i := 0; i < 100; i++ {
			for newVal := 198; newVal < 202; newVal++ {
				err := tx.Set(ch(fmt.Sprintf("key%d", i)), ch(fmt.Sprintf("val3%d", newVal)))
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	assert.NoError(t, err)

	var cachedTreeHash chainhash.Hash

	err = cachedTree.View(func(tx csmt.TreeTransactionAccess) error {
		h, err := tx.Hash()
		if err != nil {
			return err
		}
		cachedTreeHash = *h

		return nil
	})
	assert.NoError(t, err)

	err = cachedTreeDB.Flush()
	assert.NoError(t, err)

	err = underlyingTree.View(func(tx csmt.TreeTransactionAccess) error {
		h, err := tx.Hash()
		if err != nil {
			return err
		}
		underlyingHash = *h

		return nil
	})
	assert.NoError(t, err)

	assert.Equal(t, cachedTreeHash, underlyingHash)
}
