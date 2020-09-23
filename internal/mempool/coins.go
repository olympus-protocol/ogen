package mempool

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/params"

	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type coinMempoolItemMulti struct {
	transactions map[uint64]*primitives.TxMulti
	balanceSpent uint64
}

func (cmi *coinMempoolItemMulti) add(item *primitives.TxMulti, maxAmount uint64) error {
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
	cmi.transactions[txNonce] = item

	return nil
}

func newCoinMempoolItemMulti() *coinMempoolItemMulti {
	return &coinMempoolItemMulti{
		transactions: make(map[uint64]*primitives.TxMulti),
	}
}

type coinMempoolItem struct {
	transactions map[uint64]*primitives.Tx
	balanceSpent uint64
}

func (cmi *coinMempoolItem) add(item *primitives.Tx, maxAmount uint64) error {
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
	cmi.transactions[txNonce] = item

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

// CoinsMempool is an interface for coinMempool
type CoinsMempool interface {
	Add(item *primitives.Tx, state *primitives.CoinsState) error
	RemoveByBlock(b *primitives.Block)
	Get(maxTransactions uint64, s state.State) ([]*primitives.Tx, state.State)
	AddMulti(item *primitives.TxMulti, state *primitives.CoinsState) error
	GetMulti(maxTransactions uint64, s state.State) []*primitives.TxMulti
}

var _ CoinsMempool = &coinsMempool{}

// coinsMempool represents a mempool for coin transactions.
type coinsMempool struct {
	blockchain chain.Blockchain
	hostNode   hostnode.HostNode
	params     *params.ChainParams
	topic      *pubsub.Topic
	ctx        context.Context
	log        logger.Logger

	mempool      map[[20]byte]*coinMempoolItem
	mempoolMulti map[[20]byte]*coinMempoolItemMulti
	balances     map[[20]byte]uint64
	lock         sync.RWMutex
	baLock       sync.RWMutex
}

// AddMulti adds an item to the coins mempool.
func (cm *coinsMempool) AddMulti(item *primitives.TxMulti, state *primitives.CoinsState) error {
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

	mpi, ok := cm.mempoolMulti[fpkh]
	if !ok {
		cm.mempoolMulti[fpkh] = newCoinMempoolItemMulti()
		mpi = cm.mempoolMulti[fpkh]
	}
	if err := mpi.add(item, state.Balances[fpkh]); err != nil {
		return err
	}

	return nil
}

// Add adds an item to the coins mempool.
func (cm *coinsMempool) Add(item *primitives.Tx, state *primitives.CoinsState) error {
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
func (cm *coinsMempool) RemoveByBlock(b *primitives.Block) {
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
func (cm *coinsMempool) Get(maxTransactions uint64, s state.State) ([]*primitives.Tx, state.State) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	allTransactions := make([]*primitives.Tx, 0, maxTransactions)

outer:
	for _, addr := range cm.mempool {
		for _, tx := range addr.transactions {
			if err := s.ApplyTransactionSingle(tx, [20]byte{}, cm.params); err != nil {
				continue
			}
			allTransactions = append(allTransactions, tx)
			if uint64(len(allTransactions)) >= maxTransactions {
				break outer
			}
		}
	}

	// we can prioritize here, but we aren't to keep it simple
	return allTransactions, s
}

// Get gets transactions to be included in a block. Mutates state.
func (cm *coinsMempool) GetMulti(maxTransactions uint64, s state.State) []*primitives.TxMulti {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	allTransactions := make([]*primitives.TxMulti, 0, maxTransactions)

outer:
	for _, addr := range cm.mempoolMulti {
		for _, tx := range addr.transactions {
			if err := s.ApplyTransactionMulti(tx, [20]byte{}, cm.params); err != nil {
				continue
			}
			allTransactions = append(allTransactions, tx)
			if uint64(len(allTransactions)) >= maxTransactions {
				break outer
			}
		}
	}

	// we can prioritize here, but we aren't to keep it simple
	return allTransactions
}

// func (cm *CoinsMempool) modifyBalance(tx primitives.Tx, add bool) error {
// 	if add {
// 		sendingAcc := tx.Payload.FromPubkeyHash()
// 		receiving := tx.
// 		b, exist := cm.balances[]
// 	}
// }

func (cm *coinsMempool) handleSubscription(topic *pubsub.Subscription) {
	for {
		msg, err := topic.Next(cm.ctx)
		if err != nil {
			if err != cm.ctx.Err() {
				cm.log.Warnf("error getting next message in coins topic: %s", err)
				return
			}
			return
		}

		buf := bytes.NewBuffer(msg.Data)

		txMsg, err := p2p.ReadMessage(buf, cm.hostNode.GetNetMagic())
		if err != nil {
			cm.log.Warnf("unable to decode message: %s", err)
			return
		}

		tx, ok := txMsg.(*p2p.MsgTx)
		if !ok {
			cm.log.Warnf("peer sent wrong message on message subscription")
			return
		}

		currentState := cm.blockchain.State().TipState().GetCoinsState()

		err = cm.Add(tx.Data, &currentState)
		if err != nil {
			cm.log.Debugf("error adding transaction to mempool (might not be synced): %s", err)
			if err.Error() == "invalid nonce" {

			}
		}
	}
}

func (cm *coinsMempool) handleSubscriptionMulti(topic *pubsub.Subscription) {
	for {
		msg, err := topic.Next(cm.ctx)
		if err != nil {
			if err != cm.ctx.Err() {
				cm.log.Warnf("error getting next message in coins multi topic: %s", err)
				return
			}
			return
		}

		buf := bytes.NewBuffer(msg.Data)

		txMultiMsg, err := p2p.ReadMessage(buf, cm.hostNode.GetNetMagic())
		if err != nil {
			cm.log.Warnf("unable to decode message: %s", err)
			return
		}

		txMulti, ok := txMultiMsg.(*p2p.MsgTxMulti)
		if !ok {
			cm.log.Warnf("peer sent wrong message on tx multi subscription")
			return
		}

		currentState := cm.blockchain.State().TipState().GetCoinsState()

		err = cm.AddMulti(txMulti.Data, &currentState)
		if err != nil {
			cm.log.Debugf("error adding transaction to mempool (might not be synced): %s", err)
			if err.Error() == "invalid nonce" {

			}
		}
	}
}

// NewCoinsMempool constructs a new coins mempool.
func NewCoinsMempool(ctx context.Context, log logger.Logger, ch chain.Blockchain, hostNode hostnode.HostNode, params *params.ChainParams) (CoinsMempool, error) {

	topic, err := hostNode.Topic(p2p.MsgTxCmd)
	if err != nil {
		return nil, err
	}

	topicMulti, err := hostNode.Topic(p2p.MsgTxMultiCmd)
	if err != nil {
		return nil, err
	}

	topicSub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	topicMultiSub, err := topicMulti.Subscribe()
	if err != nil {
		return nil, err
	}

	cm := &coinsMempool{
		mempool:      make(map[[20]byte]*coinMempoolItem),
		mempoolMulti: make(map[[20]byte]*coinMempoolItemMulti),
		balances:     make(map[[20]byte]uint64),
		ctx:          ctx,
		blockchain:   ch,
		hostNode:     hostNode,
		params:       params,
		topic:        topic,
		log:          log,
	}

	go cm.handleSubscription(topicSub)
	go cm.handleSubscriptionMulti(topicMultiSub)
	return cm, nil
}
