package mempool

import (
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type txItem struct {
	transactions map[uint64]*primitives.Tx
	balanceSpent uint64
}

func (ti *txItem) add(item *primitives.Tx, maxAmount uint64) error {
	txNonce := item.Nonce
	txAmount := item.Amount
	txFee := item.Fee

	if txAmount+txFee+ti.balanceSpent >= maxAmount {
		return fmt.Errorf("did not add transaction spending %d with balance of %d", txAmount+txFee+ti.balanceSpent, maxAmount)
	}

	if _, ok := ti.transactions[txNonce]; ok {
		// silently accept since we already have this
		return nil
	}

	ti.balanceSpent += txAmount + txFee
	ti.transactions[txNonce] = item

	return nil
}

func (ti *txItem) removeBefore(nonce uint64) {
	for i, tx := range ti.transactions {
		if i <= nonce {
			ti.balanceSpent -= tx.Fee + tx.Amount
			delete(ti.transactions, i)
		}
	}
}

func newCoinMempoolItem() *txItem {
	return &txItem{
		transactions: make(map[uint64]*primitives.Tx),
	}
}
