package mempool

import (
	"context"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"sync"

	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"

	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type CoinsNotifee interface {
	NotifyTx(tx *primitives.Tx)
}

var (
	ErrorAccountNotOnMempool = errors.New("the account is not being tracked by the memppol")
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
	Add(item *primitives.Tx) error
	RemoveByBlock(b *primitives.Block)
	Get(maxTransactions uint64, s state.State, feeRedeemAccount [20]byte) ([]*primitives.Tx, state.State)
	GetWithoutApply() []*primitives.Tx
	GetMempoolNonce(pkh [20]byte) (uint64, error)
	AddMulti(item *primitives.TxMulti) error
	GetMulti(maxTransactions uint64, s state.State) []*primitives.TxMulti
	Notify(n CoinsNotifee)
	Unnotify(n CoinsNotifee)
}

var _ CoinsMempool = &coinsMempool{}

// coinsMempool represents a mempool for coin transactions.
type coinsMempool struct {
	chain     chain.Blockchain
	host      hostnode.HostNode
	netParams *params.ChainParams

	ctx context.Context
	log logger.Logger

	mempool    map[[20]byte]*coinMempoolItem
	lockSingle sync.Mutex

	mempoolMulti map[[20]byte]*coinMempoolItemMulti
	lockMulti    sync.Mutex

	notifeesLock sync.Mutex
	notifees     map[CoinsNotifee]struct{}

	latestNonce map[[20]byte]uint64
	lockStats   sync.Mutex
}

// AddMulti adds an item to the coins mempool.
func (cm *coinsMempool) AddMulti(item *primitives.TxMulti) error {
	cm.lockMulti.Lock()
	defer cm.lockMulti.Unlock()

	s := cm.chain.State().TipState().GetCoinsState()

	fpkh, err := item.FromPubkeyHash()
	if err != nil {
		return err
	}

	if item.Nonce != s.Nonces[fpkh]+1 {
		return errors.New("invalid nonce")
	}

	if item.Fee < 5000 {
		return errors.New("transaction doesn't include enough fee")
	}

	mpi, ok := cm.mempoolMulti[fpkh]

	if !ok {
		cm.mempoolMulti[fpkh] = newCoinMempoolItemMulti()
		mpi = cm.mempoolMulti[fpkh]
		if err := mpi.add(item, s.Balances[fpkh]); err != nil {
			return err
		}
	}

	return nil
}

// Add adds an item to the coins mempool.
func (cm *coinsMempool) Add(item *primitives.Tx) error {

	cm.lockSingle.Lock()
	defer cm.lockSingle.Unlock()

	cs := cm.chain.State().TipState().GetCoinsState()

	fpkh, err := item.FromPubkeyHash()
	if err != nil {
		return err
	}

	cm.lockStats.Lock()
	defer cm.lockStats.Unlock()

	if latestNonce, ok := cm.latestNonce[fpkh]; ok && item.Nonce < latestNonce {
		return errors.New("invalid nonce")
	}

	// Check the state for a nonce lower than the used in transaction
	if stateNonce, ok := cs.Nonces[fpkh]; ok && item.Nonce < stateNonce || !ok && item.Nonce != 1 {
		return errors.New("invalid nonce")
	}

	if item.Fee < 5000 {
		return errors.New("transaction doesn't include enough fee")
	}

	mpi, ok := cm.mempool[fpkh]

	if !ok {
		cm.mempool[fpkh] = newCoinMempoolItem()
		mpi = cm.mempool[fpkh]
		if err := mpi.add(item, cs.Balances[fpkh]); err != nil {
			return err
		}
		cm.latestNonce[fpkh] = item.Nonce
	}

	cm.notifeesLock.Lock()
	defer cm.notifeesLock.Unlock()

	for n := range cm.notifees {
		n.NotifyTx(item)
	}

	return nil
}

// RemoveByBlock removes transactions that were in an accepted block.
func (cm *coinsMempool) RemoveByBlock(b *primitives.Block) {
	cm.lockSingle.Lock()
	cm.lockMulti.Lock()
	cm.lockStats.Lock()

	defer cm.lockSingle.Unlock()
	defer cm.lockMulti.Unlock()
	defer cm.lockStats.Unlock()

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
		if tx.Nonce == cm.latestNonce[fpkh] {
			delete(cm.latestNonce, fpkh)
		}
	}
}

// Get gets transactions to be included in a block. Mutates state.
func (cm *coinsMempool) Get(maxTransactions uint64, s state.State, feeRedeemAccount [20]byte) ([]*primitives.Tx, state.State) {
	cm.lockSingle.Lock()
	defer cm.lockSingle.Unlock()

	allTransactions := make([]*primitives.Tx, 0, maxTransactions)

outer:
	for _, addr := range cm.mempool {
		for _, tx := range addr.transactions {
			if err := s.ApplyTransactionSingle(tx, feeRedeemAccount); err != nil {
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

// GetWithoutApply gets the transactions in the mempool without mutating the state
func (cm *coinsMempool) GetWithoutApply() []*primitives.Tx {
	cm.lockSingle.Lock()
	defer cm.lockSingle.Unlock()

	var txs []*primitives.Tx
	for _, addr := range cm.mempool {
		for _, tx := range addr.transactions {
			txs = append(txs, tx)
		}
	}

	return txs
}

// Get gets transactions to be included in a block. Mutates state.
func (cm *coinsMempool) GetMulti(maxTransactions uint64, s state.State) []*primitives.TxMulti {
	cm.lockMulti.Lock()
	defer cm.lockMulti.Unlock()
	allTransactions := make([]*primitives.TxMulti, 0, maxTransactions)

outer:
	for _, addr := range cm.mempoolMulti {
		for _, tx := range addr.transactions {
			if err := s.ApplyTransactionMulti(tx, [20]byte{}); err != nil {
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

func (cm *coinsMempool) handleTx(id peer.ID, msg p2p.Message) error {
	if id == cm.host.GetHost().ID() {
		return nil
	}

	data, ok := msg.(*p2p.MsgTx)
	if !ok {
		return errors.New("wrong message on tx topic")
	}

	err := cm.Add(data.Data)

	if err != nil {
		return err
	}

	return nil
}

func (cm *coinsMempool) handleTxMulti(id peer.ID, msg p2p.Message) error {
	if id == cm.host.GetHost().ID() {
		return nil
	}

	data, ok := msg.(*p2p.MsgTxMulti)
	if !ok {
		return errors.New("wrong message on txmulti topic")
	}

	err := cm.AddMulti(data.Data)

	if err != nil {
		return err
	}

	return nil
}

func (cm *coinsMempool) Notify(n CoinsNotifee) {
	cm.notifeesLock.Lock()
	defer cm.notifeesLock.Unlock()
	cm.notifees[n] = struct{}{}
}

func (cm *coinsMempool) Unnotify(n CoinsNotifee) {
	cm.notifeesLock.Lock()
	defer cm.notifeesLock.Unlock()
	delete(cm.notifees, n)
}

// GetMempoolNonce returns the latest nonce tracked by an account in mempool.
func (cm *coinsMempool) GetMempoolNonce(pkh [20]byte) (uint64, error) {
	cm.lockStats.Lock()
	defer cm.lockStats.Unlock()
	nonce, ok := cm.latestNonce[pkh]
	if !ok {
		return 0, ErrorAccountNotOnMempool
	}
	return nonce, nil
}

// NewCoinsMempool constructs a new coins mempool.
func NewCoinsMempool(ch chain.Blockchain, hostNode hostnode.HostNode) (CoinsMempool, error) {
	ctx := config.GlobalParams.Context
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	cm := &coinsMempool{
		mempool:      make(map[[20]byte]*coinMempoolItem),
		mempoolMulti: make(map[[20]byte]*coinMempoolItemMulti),
		ctx:          ctx,
		chain:        ch,
		host:         hostNode,
		netParams:    netParams,
		log:          log,
		latestNonce:  make(map[[20]byte]uint64),

		notifees: make(map[CoinsNotifee]struct{}),
	}

	if err := cm.host.RegisterTopicHandler(p2p.MsgTxCmd, cm.handleTx); err != nil {
		return nil, err
	}

	if err := cm.host.RegisterTopicHandler(p2p.MsgTxMultiCmd, cm.handleTxMulti); err != nil {
		return nil, err
	}

	return cm, nil
}
