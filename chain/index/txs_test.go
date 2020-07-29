package index_test

import (
	"testing"

	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/stretchr/testify/assert"
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
	
	assert.NoError(t, err)

	var newLoc index.TxLocator

	err = newLoc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, locators[0], newLoc)
}

func Test_IndexStoreAndFetch(t *testing.T) {

	idx, err := index.NewTxIndex("./")

	assert.NoError(t, err)

	for i, loc := range locators {
		if i < 2 {
			err = idx.SetTx(loc, accounts[0])
			
			assert.NoError(t, err)

		}
		if i >= 2 {
			err = idx.SetTx(loc, accounts[1])

			assert.NoError(t, err)
		}
	}
	acc1Txs, err := idx.GetAccountTxs(accounts[0])
	
	assert.NoError(t, err)

	acc2Txs, err := idx.GetAccountTxs(accounts[1])
	
	assert.NoError(t, err)

	acc3Txs, err := idx.GetAccountTxs(accounts[2])
	
	assert.NoError(t, err)

	assert.Equal(t, acc1Txs.Amount, uint64(2))
	
	assert.Equal(t, acc2Txs.Amount, uint64(2))

	assert.Equal(t, index.AccountTxs(index.AccountTxs{Amount:0x0, Txs:[]chainhash.Hash{}}), acc3Txs)

	loc, err := idx.GetTx(locators[0].Hash)
	
	assert.NoError(t, err)

	assert.Equal(t, loc, locators[0])
	
}
