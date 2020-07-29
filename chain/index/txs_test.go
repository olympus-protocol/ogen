package index_test

import (
	"reflect"
	"testing"

	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var accounts = map[int][20]byte{
	0: {0x1, 0x2, 0x3},
	1: {0x1, 0x2, 0x5},
	2: {0x1, 0x2, 0x10},
}

var locators = []index.TxLocator{
	{
		Hash:  chainhash.DoubleHashH([]byte("Hash-1")),
		Block: chainhash.DoubleHashH([]byte("Block-1")),
		Index: 1,
	},
	{
		Hash:  chainhash.DoubleHashH([]byte("Hash-2")),
		Block: chainhash.DoubleHashH([]byte("Block-2")),
		Index: 2,
	},
	{
		Hash:  chainhash.DoubleHashH([]byte("Hash-3")),
		Block: chainhash.DoubleHashH([]byte("Block-3")),
		Index: 1,
	},
	{
		Hash:  chainhash.DoubleHashH([]byte("Hash-4")),
		Block: chainhash.DoubleHashH([]byte("Block-4")),
		Index: 1,
	},
}

func Test_TxLocatorSerializing(t *testing.T) {
	ser, err := locators[0].Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var newLoc index.TxLocator
	err = newLoc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := reflect.DeepEqual(newLoc, locators[0])
	if !equal {
		t.Fatal("error: serialize TxLocator")
	}
}

func Test_IndexStoreAndFetch(t *testing.T) {
	idx, err := index.NewTxIndex("./")
	if err != nil {
		t.Fatal(err)
	}
	for i, loc := range locators {
		if i < 2 {
			err = idx.SetTx(loc, accounts[0])
			if err != nil {
				t.Fatal(err)
			}
		}
		if i >= 2 {
			err = idx.SetTx(loc, accounts[1])
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	acc1Txs, err := idx.GetAccountTxs(accounts[0])
	if err != nil {
		t.Fatal(err)
	}
	acc2Txs, err := idx.GetAccountTxs(accounts[1])
	if err != nil {
		t.Fatal(err)
	}
	acc3Txs, err := idx.GetAccountTxs(accounts[2])
	if err != nil {
		t.Fatal(err)
	}
	if acc1Txs.Amount != 2 {
		t.Fatal("error: wrong amount of txs tracked.")
	}
	if acc2Txs.Amount != 2 {
		t.Fatal("error: wrong amount of txs tracked.")
	}
	equal := reflect.DeepEqual(acc3Txs, index.AccountTxs{})
	if !equal {
		t.Fatal("error: error getting empty account information.")
	}
	loc, err := idx.GetTx(locators[0].Hash)
	if err != nil {
		t.Fatal(err)
	}
	equal = reflect.DeepEqual(loc, locators[0])
	if !equal {
		t.Fatal("error: locators doesn't match")
	}
}
