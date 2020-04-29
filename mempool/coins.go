package mempool

import (
	"fmt"
	"sync"

	"github.com/olympus-protocol/ogen/primitives"
)

type coinMempoolItem struct {
	transactions map[uint64]primitives.CoinPayload
	balanceSpent uint64
}

func (cmi *coinMempoolItem) add(item primitives.CoinPayload, maxAmount uint64) error {
	if item.Amount+item.Fee+cmi.balanceSpent >= maxAmount {
		return fmt.Errorf("did not add transaction spending %d with balance of %d", item.Amount+item.Fee+cmi.balanceSpent, maxAmount)
	}

	if _, ok := cmi.transactions[item.Nonce]; ok {
		return fmt.Errorf("found existing transaction with same nonce: %d", item.Nonce)
	}

	cmi.balanceSpent += item.Amount + item.Fee
	cmi.transactions[item.Nonce] = item

	return nil
}

func (cmi *coinMempoolItem) removeBefore(nonce uint64) {
	for i := range cmi.transactions {
		if i <= nonce {
			cmi.balanceSpent -= cmi.transactions[i].Fee + cmi.transactions[i].Amount
			delete(cmi.transactions, i)
		}
	}
}

func newCoinMempoolItem() *coinMempoolItem {
	return &coinMempoolItem{
		transactions: make(map[uint64]primitives.CoinPayload),
	}
}

// CoinsMempool represents a mempool for coin transactions.
type CoinsMempool struct {
	mempool map[[20]byte]*coinMempoolItem
	lock    sync.RWMutex
}

// Add adds an item to the coins mempool.
func (cm *CoinsMempool) Add(item primitives.CoinPayload, state *primitives.UtxoState) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	fpkh := item.FromPubkeyHash()
	mpi, ok := cm.mempool[fpkh]
	if !ok {
		cm.mempool[fpkh] = newCoinMempoolItem()
		mpi = cm.mempool[fpkh]
	}

	if err := mpi.add(item, state.Balances[fpkh]); err != nil {
		return err
	}

	return nil
}

// RemoveByBlock removes transactions that were in an accepted block.
func (cm *CoinsMempool) RemoveByBlock(b *primitives.Block) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	for _, tx := range b.Txs {
		switch p := tx.Payload.(type) {
		case *primitives.CoinPayload:
			fpkh := p.FromPubkeyHash()
			mempoolItem, found := cm.mempool[fpkh]
			if !found {
				continue
			}
			mempoolItem.removeBefore(p.Nonce)
			if mempoolItem.balanceSpent == 0 {
				delete(cm.mempool, fpkh)
			}
		}
	}
}

// Get gets transactions to be included in a block.
func (cm *CoinsMempool) Get(maxTransactions uint64, state primitives.State) []primitives.Tx {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	allTransactions := make([]primitives.Tx, 0, maxTransactions)

	stateTest := state.Copy()

outer:
	for _, addr := range cm.mempool {
		for _, tx := range addr.transactions {
			if err := stateTest.UtxoState.ApplyTransaction(&tx, [20]byte{}); err != nil {
				continue
			}
			allTransactions = append(allTransactions, primitives.Tx{
				TxVersion: 0,
				TxType:    primitives.TxCoins,
				Payload:   &tx,
			})
			if uint64(len(allTransactions)) >= maxTransactions {
				break outer
			}
		}
	}

	// we can prioritize here, but we aren't to keep it simple
	return allTransactions
}

// NewCoinsMempool constructs a new coins mempool.
func NewCoinsMempool() *CoinsMempool {
	return &CoinsMempool{
		mempool: make(map[[20]byte]*coinMempoolItem),
	}
}
