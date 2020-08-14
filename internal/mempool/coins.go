package mempool

import (
	"context"
	"errors"
	"fmt"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/peers"
	"github.com/olympus-protocol/ogen/pkg/params"

	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type coinMempoolItem struct {
	transactions map[uint64]*primitives.Tx
	balanceSpent uint64
}

func (cmi *coinMempoolItem) add(item primitives.Tx, maxAmount uint64) error {
	txNonce := item.Nonce
	txAmount := item.Amount
	txFee := item.Fee

	if txAmount+txFee+cmi.balanceSpent >= maxAmount {
		return fmt.Errorf("did not add transaction spending %d with balance of %d", txAmount+txFee+cmi.balanceSpent, maxAmount)
	}

	if _, ok := cmi.transactions[txNonce]; ok {
		// silently accept since we already have this
		return nil
	}

	cmi.balanceSpent += txAmount + txFee
	cmi.transactions[txNonce] = &item

	return nil
}

func (cmi *coinMempoolItem) removeBefore(nonce uint64) {
	for i, tx := range cmi.transactions {
		if i <= nonce {
			cmi.balanceSpent -= tx.Fee + tx.Amount
			delete(cmi.transactions, i)
		}
	}
}

func newCoinMempoolItem() *coinMempoolItem {
	return &coinMempoolItem{
		transactions: make(map[uint64]*primitives.Tx),
	}
}

// CoinsMempool represents a mempool for coin transactions.
type CoinsMempool struct {
	blockchain chain.Blockchain
	hostNode   peers.HostNode
	params     *params.ChainParams
	topic      *pubsub.Topic
	ctx        context.Context
	log        *logger.Logger

	mempool  map[[20]byte]*coinMempoolItem
	balances map[[20]byte]uint64
	lock     sync.RWMutex
	baLock   sync.RWMutex
}

// Add adds an item to the coins mempool.
func (cm *CoinsMempool) Add(item primitives.Tx, state *primitives.CoinsState) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	fpkh, err := item.FromPubkeyHash()
	if err != nil {
		return err
	}

	if item.Nonce != state.Nonces[fpkh]+1 {
		return errors.New("invalid nonce")
	}

	if item.Fee < 5000 {
		return errors.New("transaction doesn't include enough fee")
	}

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
		fpkh, err := tx.FromPubkeyHash()
		if err != nil {
			continue
		}
		mempoolItem, found := cm.mempool[fpkh]
		if !found {
			continue
		}
		mempoolItem.removeBefore(tx.Nonce)
		if mempoolItem.balanceSpent == 0 {
			delete(cm.mempool, fpkh)
		}
	}
}

// Get gets transactions to be included in a block. Mutates state.
func (cm *CoinsMempool) Get(maxTransactions uint64, state *primitives.State) ([]*primitives.Tx, *primitives.State) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	allTransactions := make([]*primitives.Tx, 0, maxTransactions)

outer:
	for _, addr := range cm.mempool {
		for _, tx := range addr.transactions {
			if err := state.ApplyTransactionSingle(tx, [20]byte{}, cm.params); err != nil {
				continue
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

			cm.log.Warnf("peer sent invalid transaction: %s", err)
			err = cm.hostNode.BanScorePeer(msg.GetFrom(), 100)
			if err == nil {
				cm.log.Warnf("peer %s was banned", msg.GetFrom().String())
			}
			continue
		}

		currentState := cm.blockchain.State().TipState().CoinsState

		err = cm.Add(*tx, &currentState)
		if err != nil {
			cm.log.Debugf("error adding transaction to mempool (might not be synced): %s", err)
			if err.Error() == "invalid nonce" {
				err = cm.hostNode.BanScorePeer(msg.GetFrom(), 10)
				if err == nil {
					cm.log.Warnf("peer %s banscore was increased", msg.GetFrom().String())
				}
			}
		}
	}
}

// NewCoinsMempool constructs a new coins mempool.
func NewCoinsMempool(ctx context.Context, log *logger.Logger, ch chain.Blockchain, hostNode peers.HostNode, params *params.ChainParams) (*CoinsMempool, error) {
	topic, err := hostNode.Topic("tx")
	if err != nil {
		return nil, err
	}

	topicSub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	_, err = topic.Relay()
	if err != nil {
		return nil, err
	}

	cm := &CoinsMempool{
		mempool:    make(map[[20]byte]*coinMempoolItem),
		balances:   make(map[[20]byte]uint64),
		ctx:        ctx,
		blockchain: ch,
		hostNode:   hostNode,
		params:     params,
		topic:      topic,
		log:        log,
	}

	go cm.handleSubscription(topicSub)

	return cm, nil
}
