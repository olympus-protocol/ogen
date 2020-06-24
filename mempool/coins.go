package mempool

import (
	"context"
	"fmt"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/utils/logger"

	"github.com/olympus-protocol/ogen/primitives"
)

type coinMempoolItem struct {
	transactions map[uint64]primitives.Tx
	balanceSpent uint64
}

func (cmi *coinMempoolItem) add(item primitives.Tx, maxAmount uint64) error {
	txNonce := item.Payload.GetNonce()
	txAmount := item.Payload.GetAmount()
	txFee := item.Payload.GetFee()

	if txAmount+txFee+cmi.balanceSpent >= maxAmount {
		return fmt.Errorf("did not add transaction spending %d with balance of %d", txAmount+txFee+cmi.balanceSpent, maxAmount)
	}

	if _, ok := cmi.transactions[txNonce]; ok {
		// silently accept since we already have this
		return nil
	}

	cmi.balanceSpent += txAmount + txFee
	cmi.transactions[txNonce] = item

	return nil
}

func (cmi *coinMempoolItem) removeBefore(nonce uint64) {
	for i := range cmi.transactions {
		if i <= nonce {
			cmi.balanceSpent -= cmi.transactions[i].Payload.GetFee() + cmi.transactions[i].Payload.GetAmount()
			delete(cmi.transactions, i)
		}
	}
}

func newCoinMempoolItem() *coinMempoolItem {
	return &coinMempoolItem{
		transactions: make(map[uint64]primitives.Tx),
	}
}

// CoinsMempool represents a mempool for coin transactions.
type CoinsMempool struct {
	blockchain *chain.Blockchain
	params     *params.ChainParams
	topic      *pubsub.Topic
	ctx        context.Context
	log        *logger.Logger

	mempool  map[[20]byte]*coinMempoolItem
	balances map[[20]byte]uint64
	lock     sync.RWMutex
	baLock   sync.RWMutex
}

// func (cm *CoinsMempool) GetVotesNotInBloom(bloom *bloom.BloomFilter) []primitives.CoinPayload {
// 	cm.lock.RLock()
// 	defer cm.lock.RUnlock()
// 	txs := make([]primitives.CoinPayload, 0)
// 	for _, item := range cm.mempool {
// 		for _, tx := range item.transactions {
// 			vh := tx.Hash()
// 			if bloom.Has(vh) {
// 				continue
// 			}

// 			txs = append(txs, tx)
// 		}
// 	}
// 	return txs
// }

// Add adds an item to the coins mempool.
func (cm *CoinsMempool) Add(item primitives.Tx, state *primitives.CoinsState) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	fpkh := item.Payload.FromPubkeyHash()
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
		case *primitives.TransferSinglePayload:
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

// Get gets transactions to be included in a block. Mutates state.
func (cm *CoinsMempool) Get(maxTransactions uint64, state *primitives.State) ([]primitives.Tx, *primitives.State) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	allTransactions := make([]primitives.Tx, 0, maxTransactions)

outer:
	for _, addr := range cm.mempool {
		for _, tx := range addr.transactions {
			switch u := tx.Payload.(type) {
			case *primitives.TransferSinglePayload:
				if err := state.ApplyTransactionSingle(u, [20]byte{}, cm.params); err != nil {
					continue
				}
			case *primitives.TransferMultiPayload:
				if err := state.ApplyTransactionMulti(u, [20]byte{}, cm.params); err != nil {
					continue
				}
			}

			allTransactions = append(allTransactions, tx)
			if uint64(len(allTransactions)) >= maxTransactions {
				break outer
			}
		}
	}

	// we can prioritize here, but we aren't to keep it simple
	return allTransactions, state
}

// func (cm *CoinsMempool) modifyBalance(tx primitives.Tx, add bool) error {
// 	if add {
// 		sendingAcc := tx.Payload.FromPubkeyHash()
// 		receiving := tx.
// 		b, exist := cm.balances[]
// 	}
// }

func (cm *CoinsMempool) handleSubscription(topic *pubsub.Subscription) {
	for {
		msg, err := topic.Next(cm.ctx)
		if err != nil {
			cm.log.Warnf("error getting next message in coins topic: %s", err)
			return
		}

		tx := new(primitives.Tx)

		if err := tx.Unmarshal(msg.Data); err != nil {
			// TODO: ban peer
			cm.log.Warnf("peer sent invalid transaction: %s", err)
			continue
		}

		currentState := cm.blockchain.State().TipState().CoinsState

		err = cm.Add(*tx, &currentState)
		if err != nil {
			cm.log.Debugf("error adding transaction to mempool (might not be synced): %s", err)
		}
	}
}

// NewCoinsMempool constructs a new coins mempool.
func NewCoinsMempool(ctx context.Context, log *logger.Logger, ch *chain.Blockchain, hostNode *peers.HostNode, params *params.ChainParams) (*CoinsMempool, error) {
	topic, err := hostNode.Topic("tx")
	if err != nil {
		return nil, err
	}

	topicSub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cm := &CoinsMempool{
		mempool:    make(map[[20]byte]*coinMempoolItem),
		balances:   make(map[[20]byte]uint64),
		ctx:        ctx,
		blockchain: ch,
		params:     params,
		topic:      topic,
		log:        log,
	}

	go cm.handleSubscription(topicSub)

	return cm, nil
}
