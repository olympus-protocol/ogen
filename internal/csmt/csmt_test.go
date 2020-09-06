package csmt_test

import (
	"fmt"
	"github.com/olympus-protocol/ogen/internal/csmt"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"reflect"
	"testing"
)

func ch(s string) chainhash.Hash {
	return chainhash.HashH([]byte(s))
}

func TestTree_RandomSet(t *testing.T) {
	keys := make([]chainhash.Hash, 500)
	val := ch("testval")
	tree := csmt.NewTree(csmt.NewInMemoryTreeDB())

	for i := range keys {
		keys[i] = ch(fmt.Sprintf("%d", i))
	}

	err := tree.Update(func(tx csmt.TreeTransactionAccess) error {
		for i := 0; i < 500; i++ {
			err := tx.Set(keys[i], val)
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

func TestTree_SetZero(t *testing.T) {
	val := chainhash.Hash{}

	tree := csmt.NewTree(csmt.NewInMemoryTreeDB())

	err := tree.Update(func(tx csmt.TreeTransactionAccess) error {
		err := tx.Set(ch("1"), val)
		if err != nil {
			return err
		}
		err = tx.Set(ch("2"), val)
		if err != nil {
			return err
		}
		err = tx.Set(ch("3"), val)
		if err != nil {
			return err
		}
		err = tx.Set(ch("4"), val)
		if err != nil {
			return err
		}
		err = tx.Set(ch("5"), val)
		if err != nil {
			return err
		}
		th, err := tx.Hash()
		if err != nil {
			return err
		}

		if !th.IsEqual(&primitives.EmptyTrees[255]) {
			return fmt.Errorf("expected tree to match %s but got %s", primitives.EmptyTrees[255], th)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkTree_Set(b *testing.B) {
	keys := make([]chainhash.Hash, b.N)
	val := ch("testval")
	t := csmt.NewTree(csmt.NewInMemoryTreeDB())

	for i := range keys {
		keys[i] = ch(fmt.Sprintf("%d", i))
	}

	b.ResetTimer()

	err := t.Update(func(tx csmt.TreeTransactionAccess) error {
		for i := 0; i < b.N; i++ {
			err := tx.Set(keys[i], val)
			if err != nil {
				b.Fatal(err)
			}
		}
		return nil
	})
	if err != nil {
		b.Fatal(err)
	}
}

func Test_calculateSubtreeHashWithOneLeaf(t *testing.T) {
	type args struct {
		key     chainhash.Hash
		value   chainhash.Hash
		atLevel uint8
	}
	tests := []struct {
		name string
		args args
		want chainhash.Hash
	}{
		{
			name: "test lowest node",
			args: args{
				key:     ch("test"),
				value:   chainhash.Hash{},
				atLevel: 0,
			},
			want: chainhash.Hash{},
		},
		{
			name: "test empty subtree root",
			args: args{
				key:     ch("test"),
				value:   chainhash.Hash{},
				atLevel: 255,
			},
			want: primitives.EmptyTrees[255],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := csmt.CalculateSubtreeHashWithOneLeaf(&tt.args.key, &tt.args.value, tt.args.atLevel); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateSubtreeHashWithOneLeaf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandomGenerateUpdateWitness(t *testing.T) {
	keys := make([]chainhash.Hash, 500)
	val := ch("testval")
	treeDB := csmt.NewInMemoryTreeDB()
	tree := csmt.NewTree(treeDB)

	for i := range keys {
		keys[i] = ch(fmt.Sprintf("%d", i))
	}

	var treehash *chainhash.Hash

	err := tree.Update(func(tx csmt.TreeTransactionAccess) error {
		for i := 0; i < 500; i++ {
			err := tx.Set(keys[i], val)
			if err != nil {
				t.Fatal(err)
			}
		}

		h, err := tx.Hash()
		if err != nil {
			t.Fatal(err)
		}

		treehash = h
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1; i++ {
		err := treeDB.View(func(tx csmt.TreeDatabaseTransaction) error {
			w, err := csmt.GenerateUpdateWitness(tx, keys[i], val)
			if err != nil {
				return err
			}
			root, err := csmt.CalculateRoot(keys[i], val, w.WitnessBitfield, w.Witnesses, w.LastLevel)
			if err != nil {
				t.Fatal(err)
			}
			if !root.IsEqual(treehash) {
				t.Fatalf("expected witness root to equal tree hash (expected: %s, got: %s)", treehash, root)
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestGenerateUpdateWitnessEmptyTree(t *testing.T) {
	treeDB := csmt.NewInMemoryTreeDB()
	tree := csmt.NewTree(treeDB)

	var uw *primitives.UpdateWitness
	err := treeDB.View(func(tx csmt.TreeDatabaseTransaction) error {
		w, err := csmt.GenerateUpdateWitness(tx, ch("asdf"), ch("1"))
		uw = w
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	var th *chainhash.Hash
	err = tree.View(func(tx csmt.TreeTransactionAccess) error {
		h, err := tx.Hash()
		if err != nil {
			t.Fatal(err)
		}

		th = h
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	newRoot, err := csmt.ApplyWitness(*uw, *th)
	if err != nil {
		t.Fatal(err)
	}

	err = tree.Update(func(tx csmt.TreeTransactionAccess) error {
		err = tx.Set(ch("asdf"), ch("1"))
		if err != nil {
			return err
		}

		th, err = tx.Hash()
		if err != nil {
			return err
		}
		if !th.IsEqual(newRoot) {
			return fmt.Errorf("expected calculated state root (%s) to match tree state root (%s)", newRoot, th)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerateUpdateWitnessUpdate(t *testing.T) {
	treeDB := csmt.NewInMemoryTreeDB()
	tree := csmt.NewTree(treeDB)

	err := tree.Update(func(tx csmt.TreeTransactionAccess) error {
		err := tx.Set(ch("asdf"), ch("2"))
		if err != nil {
			return err
		}
		err = tx.Set(ch("asdf1"), ch("2"))
		if err != nil {
			return err
		}
		err = tx.Set(ch("asdf2"), ch("2"))
		if err != nil {
			return err
		}
		err = tx.Set(ch("asdf3"), ch("2"))
		if err != nil {
			return err
		}
		err = tx.Set(ch("asdf4"), ch("2"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1; i++ {
		setVal := fmt.Sprintf("%d", i)

		var uw *primitives.UpdateWitness

		err := treeDB.Update(func(tx csmt.TreeDatabaseTransaction) error {
			w, err := csmt.GenerateUpdateWitness(tx, ch("asdf"), ch(setVal))
			uw = w
			return err
		})

		if err != nil {
			t.Fatal(err)
		}

		th, err := tree.Hash()
		if err != nil {
			t.Fatal(err)
		}

		newRoot, err := csmt.ApplyWitness(*uw, th)
		if err != nil {
			t.Fatal(err)
		}

		err = tree.Update(func(tx csmt.TreeTransactionAccess) error {
			return tx.Set(ch("asdf"), ch(setVal))
		})
		if err != nil {
			t.Fatal(err)
		}

		th, err = tree.Hash()
		if err != nil {
			t.Fatal(err)
		}
		if !th.IsEqual(newRoot) {
			t.Fatalf("expected calculated state root (%s) to match tree state root (%s)", newRoot, th)
		}
	}
}

func BenchmarkGenerateUpdateWitness(b *testing.B) {
	keys := make([]chainhash.Hash, b.N)
	val := ch("testval")
	treeDB := csmt.NewInMemoryTreeDB()
	tree := csmt.NewTree(treeDB)

	for i := range keys {
		keys[i] = ch(fmt.Sprintf("%d", i))
	}

	err := tree.Update(func(tx csmt.TreeTransactionAccess) error {
		for i := 0; i < b.N; i++ {
			err := tx.Set(keys[i], val)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	err = treeDB.View(func(tx csmt.TreeDatabaseTransaction) error {
		for i := 0; i < b.N; i++ {
			_, err := csmt.GenerateUpdateWitness(tx, keys[i], val)
			if err != nil {
				b.Fatal(err)
			}
		}

		return nil
	})
	if err != nil {
		b.Fatal(err)
	}
}

func TestChainedUpdates(t *testing.T) {
	tree := csmt.NewTree(csmt.NewInMemoryTreeDB())

	initialRoot, err := tree.Hash()
	if err != nil {
		t.Fatal(err)
	}
	witnesses := make([]*primitives.UpdateWitness, 0)

	err = tree.Update(func(txA csmt.TreeTransactionAccess) error {
		tx := txA.(*csmt.TreeTransaction)
		// start by generating a bunch of witnesses
		for i := 0; i < 1000; i++ {
			key := ch(fmt.Sprintf("key%d", i))

			// test empty key
			testProof2, err := tx.Prove(key)
			if err != nil {
				return err
			}

			th, err := tx.Hash()
			if err != nil {
				return err
			}
			if csmt.CheckWitness(testProof2, *th) == false {
				return fmt.Errorf("expected verification witness to verify")
			}

			val := ch(fmt.Sprintf("val%d", i))

			uw, err := tx.SetWithWitness(key, val)
			if err != nil {
				return err
			}
			witnesses = append(witnesses, uw)

			testProof, err := tx.Prove(key)
			if err != nil {
				return err
			}
			th, err = tx.Hash()
			if err != nil {
				return err
			}
			if csmt.CheckWitness(testProof, *th) == false {
				return fmt.Errorf("expected verification witness to verify")
			}
		}

		// then update half of them
		for i := 0; i < 500; i++ {
			key := ch(fmt.Sprintf("key%d", i))

			// test empty key
			testProof2, err := tx.Prove(key)
			if err != nil {
				return err
			}
			th, err := tx.Hash()
			if err != nil {
				return err
			}
			if csmt.CheckWitness(testProof2, *th) == false {
				return fmt.Errorf("expected verification witness to verify")
			}

			val := ch(fmt.Sprintf("val1%d", i))

			uw, err := tx.SetWithWitness(key, val)
			if err != nil {
				return err
			}

			witnesses = append(witnesses, uw)

			testProof, err := tx.Prove(key)
			if err != nil {
				return err
			}
			th, err = tx.Hash()
			if err != nil {
				return err
			}
			if csmt.CheckWitness(testProof, *th) == false {
				return fmt.Errorf("expected verification witness to verify")
			}
		}

		currentRoot := initialRoot
		for i := range witnesses {
			newRoot, err := csmt.ApplyWitness(*witnesses[i], currentRoot)
			if err != nil {
				return err
			}
			currentRoot = *newRoot
		}

		th, err := tx.Hash()
		if err != nil {
			return err
		}
		if !th.IsEqual(&currentRoot) {
			return fmt.Errorf("expected hash after applying updates to match")
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEmptyBranchWitness(t *testing.T) {
	tree := csmt.NewTree(csmt.NewInMemoryTreeDB())

	err := tree.Update(func(txA csmt.TreeTransactionAccess) error {
		tx := txA.(*csmt.TreeTransaction)

		preroot, err := tx.Hash()
		if err != nil {
			t.Fatal(err)
		}

		w0, err := tx.SetWithWitness(ch("test"), ch("asdf"))
		if err != nil {
			return err
		}

		w1, err := tx.SetWithWitness(ch("asdfghi"), ch("asdf"))
		if err != nil {
			return err
		}

		newRoot, err := csmt.ApplyWitness(*w0, *preroot)
		if err != nil {
			return err
		}

		newRoot, err = csmt.ApplyWitness(*w1, *newRoot)
		return err
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckWitness(t *testing.T) {
	tree := csmt.NewTree(csmt.NewInMemoryTreeDB())
	//preroot := tree.Hash()

	err := tree.Update(func(txA csmt.TreeTransactionAccess) error {
		tx := txA.(*csmt.TreeTransaction)

		err := tx.Set(ch("test"), ch("asdf"))
		if err != nil {
			return err
		}
		err = tx.Set(ch("asdfghi"), ch("asdf"))
		if err != nil {
			return err
		}

		testProof, err := tx.Prove(ch("test"))
		if err != nil {
			return err
		}
		th, err := tx.Hash()
		if err != nil {
			return err
		}
		if csmt.CheckWitness(testProof, *th) == false {
			t.Fatal("expected verification witness to verify")
		}

		// test empty key
		testProof2, err := tx.Prove(ch("test1"))
		if err != nil {
			return err
		}
		th, err = tx.Hash()
		if err != nil {
			return err
		}
		if csmt.CheckWitness(testProof2, *th) == false {
			t.Fatal("expected verification witness to verify")
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

}
