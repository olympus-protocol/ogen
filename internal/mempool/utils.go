package mempool

import (
	"encoding/binary"
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type PoolType uint64

const (
	PoolTypeDeposit PoolType = iota
	PoolTypeExit
	PoolTypePartialExit
	PoolTypeLatestNonce
	PoolTypeGovernanceVote
	PoolTypeCoinProof
	PoolTypeVote
)

func appendKey(k []byte, t PoolType) []byte {
	var key []byte
	switch t {
	case PoolTypeDeposit:
		key = append(key, []byte("deposit-")...)
		key = append(key, k...)
		return key
	case PoolTypeExit:
		key = append(key, []byte("exit-")...)
		key = append(key, k...)
		return key
	case PoolTypePartialExit:
		key = append(key, []byte("partial_exit-")...)
		key = append(key, k...)
		return key
	case PoolTypeLatestNonce:
		key = append(key, []byte("latest_nonce-")...)
		key = append(key, k...)
		return key
	case PoolTypeGovernanceVote:
		key = append(key, []byte("governance_vote-")...)
		key = append(key, k...)
		return key
	case PoolTypeCoinProof:
		key = append(key, []byte("coin_proof-")...)
		key = append(key, k...)
		return key
	case PoolTypeVote:
		key = append(key, []byte("vote-")...)
		key = append(key, k...)
		return key
	default:
		return k
	}
}

func nonceToBytes(nonce uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, nonce)
	return b
}

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
