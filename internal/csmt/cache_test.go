package csmt_test

import (
	"fmt"
	"github.com/olympus-protocol/ogen/internal/csmt"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"testing"
)

func TestRandomWritesRollbackCommit(t *testing.T) {
	under := csmt.NewInMemoryTreeDB()

	underlyingTree := csmt.NewTree(under)

	err := underlyingTree.Update(func(tx csmt.TreeTransactionAccess) error {
		for i := 0; i < 200; i++ {
			err := tx.Set(ch(fmt.Sprintf("key%d", i)), ch(fmt.Sprintf("val%d", i)))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	treeRoot, err := underlyingTree.Hash()
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
			for i := 198; i < 202; i++ {
				err := tx.Set(ch(fmt.Sprintf("key%d", i)), ch(fmt.Sprintf("val2%d", i)))
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	underlyingHash, err := underlyingTree.Hash()
	if err != nil {
		t.Fatal(err)
	}

	if !underlyingHash.IsEqual(&treeRoot) {
		t.Fatal("expected uncommitted transaction not to affect underlying tree")
	}

	for i := 0; i < 100; i++ {
		cachedTreeDB, err := csmt.NewTreeMemoryCache(under)
		if err != nil {
			t.Fatal(err)
		}
		cachedTree := csmt.NewTree(cachedTreeDB)

		err = cachedTree.Update(func(tx csmt.TreeTransactionAccess) error {
			for newVal := 198; newVal < 202; newVal++ {
				err := tx.Set(ch(fmt.Sprintf("key%d", i)), ch(fmt.Sprintf("val3%d", newVal)))
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		cachedTreeHash, err := cachedTree.Hash()
		if err != nil {
			t.Fatal(err)
		}

		err = cachedTreeDB.Flush()
		if err != nil {
			t.Fatal(err)
		}

		underlyingHash, err := underlyingTree.Hash()
		if err != nil {
			t.Fatal(err)
		}

		if !cachedTreeHash.IsEqual(&underlyingHash) {
			t.Fatal("expected flush to update the underlying tree")
		}
	}

	setNodeHashes := make(map[chainhash.Hash]struct{})

	err = under.View(func(tx csmt.TreeDatabaseTransaction) error {
		root, _ := tx.Root()

		queue := []*csmt.Node{root}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if current == nil {
				continue
			}

			setNodeHashes[current.GetHash()] = struct{}{}

			if current.Right() != nil {
				right, _ := tx.GetNode(*current.Right())
				queue = append(queue, right)
			}

			if current.Left() != nil {
				left, _ := tx.GetNode(*current.Left())
				queue = append(queue, left)
			}
		}

		for nodeHash := range under.Nodes() {
			if _, found := setNodeHashes[nodeHash]; !found {
				return fmt.Errorf("did not clean up node with hash %s", nodeHash)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
