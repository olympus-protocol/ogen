package primitives

import "sync"

type CoinsState struct {
	balanceLock  sync.RWMutex
	nonceLocks   sync.RWMutex
	balanceIndex map[[20]byte]int
	nonceIndex   map[[20]byte]int
	Balances     []uint64
	Nonces       []uint64
}

func NewCoinsBalances() CoinsState {
	return CoinsState{nonceIndex: make(map[[20]byte]int), balanceIndex: make(map[[20]byte]int), Balances: []uint64{}, Nonces: []uint64{}}
}

func (cb *CoinsState) GetNonce(acc [20]byte) uint64 {
	i, ok := cb.getBalanceIndex(acc)
	if !ok {
		// Create new nonce and append to the index
	}
	return cb.Nonces[i]
}

func (cb *CoinsState) Get(acc [20]byte) uint64 {
	i, ok := cb.getBalanceIndex(acc)
	if !ok {
		return 0
	}
	return cb.Balances[i]
}

func (cb *CoinsState) Reduce(acc [20]byte, amount uint64) {
	i, ok := cb.getBalanceIndex(acc)
	if !ok {
		// Reduce the balance of a non-existing account (this should never happen)
		return
	}
	cb.Balances[i] -= amount
	return
}

func (cb *CoinsState) Increase(acc [20]byte, amount uint64) uint64 {
	i, ok := cb.getBalanceIndex(acc)
	if !ok {

	}
	return cb.Balances[i]
}

func (cb *CoinsState) getBalanceIndex(acc [20]byte) (int, bool) {
	cb.balanceLock.Lock()
	defer cb.balanceLock.Unlock()
	balanceIndex, ok := cb.balanceIndex[acc]
	return balanceIndex, ok
}

func (cb *CoinsState) setBalanceIndex(acc [20]byte) {
	cb.balanceLock.Lock()
	defer cb.balanceLock.Unlock()
	newIndex := len(cb.Balances)
	cb.balanceIndex[acc] = newIndex
	return
}

func (cb *CoinsState) getNonceIndex(acc [20]byte) (int, bool) {
	cb.nonceLocks.Lock()
	defer cb.nonceLocks.Unlock()
	nonceIndex, ok := cb.nonceIndex[acc]
	return nonceIndex, ok
}
