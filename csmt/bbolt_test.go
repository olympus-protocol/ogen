package csmt_test

import (
	"fmt"
	"testing"

	"github.com/olympus-protocol/ogen/csmt"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"go.etcd.io/bbolt"
)

func ch(s string) chainhash.Hash {
	return chainhash.HashH([]byte(s))
}

func TestRandomWritesRollbackCommitBolt(t *testing.T) {
	bboltdb, err := bbolt.Open("db.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = bboltdb.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket([]byte("test_bucket"))
	})

	defer bboltdb.Close()

	under := csmt.NewBoltTreeDB(bboltdb, "test_bucket")

	underlyingTree := csmt.NewTree(under)

	var treeRoot chainhash.Hash

	err = underlyingTree.Update(func(tx csmt.TreeTransactionAccess) error {
		for i := 0; i < 200; i++ {
			err := tx.Set(ch(fmt.Sprintf("key%d", i)), ch(fmt.Sprintf("val%d", i)))
			fmt.Println(err)
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

	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		cachedTreeDB, err := csmt.NewTreeMemoryCache(under)
		if err != nil {
			t.Fatal(err)
		}
		cachedTree := csmt.NewTree(cachedTreeDB)

		err = cachedTree.Update(func(tx csmt.TreeTransactionAccess) error {
			for newVal := 198; newVal < 202; newVal++ {
				err := tx.Set(ch(fmt.Sprintf("key%d", i)), ch(fmt.Sprintf("val2%d", newVal)))
				if err != nil {
					t.Fatal(err)
				}
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
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
	if err != nil {
		t.Fatal(err)
	}

	if !underlyingHash.IsEqual(&treeRoot) {
		t.Fatal("expected uncommitted transaction not to affect underlying tree")
	}

	cachedTreeDB, err := csmt.NewTreeMemoryCache(under)
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}

	var cachedTreeHash chainhash.Hash

	err = cachedTree.View(func(tx csmt.TreeTransactionAccess) error {
		h, err := tx.Hash()
		if err != nil {
			return err
		}
		cachedTreeHash = *h

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = cachedTreeDB.Flush()
	if err != nil {
		t.Fatal(err)
	}

	err = underlyingTree.View(func(tx csmt.TreeTransactionAccess) error {
		h, err := tx.Hash()
		if err != nil {
			return err
		}
		underlyingHash = *h

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if !cachedTreeHash.IsEqual(&underlyingHash) {
		t.Fatal("expected flush to update the underlying tree")
	}
}
