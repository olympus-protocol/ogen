package primitives

import (
	"sync"
)

// AccountInfo is the information contained into both slices. It represents the account hash and a value.
type AccountInfo struct {
	Account [20]byte
	Info    uint64
}

var (
	balanceLock  sync.RWMutex
	nonceLocks   sync.RWMutex
	balanceIndex = map[[20]byte]int{}
	nonceIndex   = map[[20]byte]int{}
)

// CoinsState is the serializable struct with the access indexes for fast fetch balances and nonces.
type CoinsState struct {
	Balances []AccountInfo
	Nonces   []AccountInfo
}

// Load is used when the Balances and Nonces slices are already filled. This constructs the index maps
func (cs *CoinsState) Load() {
	for i, v := range cs.Balances {
		balanceLock.Lock()
		balanceIndex[v.Account] = i
		balanceLock.Unlock()
	}
	for i, v := range cs.Nonces {
		nonceLocks.Lock()
		nonceIndex[v.Account] = i
		nonceLocks.Unlock()
	}
}

// NewCoinsStates is used only to initialize the coin state when chain is not synced.
func NewCoinsStates() CoinsState {
	return CoinsState{Balances: []AccountInfo{}, Nonces: []AccountInfo{}}
}

// GetTotal sums all the state balances.
func (cs *CoinsState) GetTotal() uint64 {
	total := uint64(0)
	for _, b := range cs.Balances {
		total += b.Info
	}
	return total
}

// GetNonce returns the account nonce.
func (cs *CoinsState) GetNonce(acc [20]byte) uint64 {
	i, ok := cs.getNonceIndex(acc)
	if !ok {
		return 0
	}
	return cs.Nonces[i].Info
}

// SetNonce sets the nonce to a new value.
func (cs *CoinsState) SetNonce(acc [20]byte, value uint64) {
index:
	i, ok := cs.getNonceIndex(acc)
	if !ok {
		lastIndex := len(cs.Nonces)
		cs.Nonces = append(cs.Nonces, AccountInfo{Account: acc, Info: value})
		cs.setNonceIndex(acc, lastIndex)
		goto index
	}
	cs.Nonces[i].Info = value
	return
}

// GetBalance returns the account balance.
func (cs *CoinsState) GetBalance(acc [20]byte) uint64 {
	i, ok := cs.getBalanceIndex(acc)
	if !ok {
		return 0
	}
	return cs.Balances[i].Info
}

// ReduceBalance reduces the account balance.
func (cs *CoinsState) ReduceBalance(acc [20]byte, amount uint64) {
	i, ok := cs.getBalanceIndex(acc)
	if !ok {
		// Reduce the balance of a non-existing account (this should never happen)
		return
	}
	cs.Balances[i].Info -= amount
	return
}

// IncreaseBalance increases the account balance.
func (cs *CoinsState) IncreaseBalance(acc [20]byte, amount uint64) {
index:
	i, ok := cs.getBalanceIndex(acc)
	if !ok {
		// Append to the slice and add to the map
		lastIndex := len(cs.Balances)
		cs.Balances = append(cs.Balances, AccountInfo{Account: acc, Info: amount})
		cs.setBalanceIndex(acc, lastIndex)
		goto index
	}
	cs.Balances[i].Info += amount
	return
}

func (cs *CoinsState) setNonceIndex(acc [20]byte, index int) bool {
	nonceLocks.Lock()
	defer nonceLocks.Unlock()
	nonceIndex[acc] = index
	return true
}

func (cs *CoinsState) setBalanceIndex(acc [20]byte, index int) bool {
	balanceLock.Lock()
	defer balanceLock.Unlock()
	balanceIndex[acc] = index
	return true
}

func (cs *CoinsState) getBalanceIndex(acc [20]byte) (int, bool) {
	balanceLock.Lock()
	defer balanceLock.Unlock()
	i, ok := balanceIndex[acc]
	return i, ok
}

func (cs *CoinsState) getNonceIndex(acc [20]byte) (int, bool) {
	nonceLocks.Lock()
	defer nonceLocks.Unlock()
	i, ok := nonceIndex[acc]
	return i, ok
}
