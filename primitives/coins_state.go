package primitives

import "sync"

// AccountInfo is the information contained into both slices. It represents the account hash and a value.
type AccountInfo struct {
	Account [20]byte
	Info    uint64
}

// CoinsState is the serializable struct with the access indexes for fast fetch balances and nonces.
type CoinsState struct {
	balanceLock  sync.RWMutex
	nonceLocks   sync.RWMutex
	balanceIndex map[[20]byte]int
	nonceIndex   map[[20]byte]int
	Balances     []AccountInfo
	Nonces       []AccountInfo
}

// Load is used when the Balances and Nonces slices are already filled. This constructs the index maps
func (cs *CoinsState) Load() {
	for i, v := range cs.Balances {
		cs.balanceLock.Lock()
		cs.balanceIndex[v.Account] = i
		cs.balanceLock.Unlock()
	}
	for i, v := range cs.Nonces {
		cs.nonceLocks.Lock()
		cs.nonceIndex[v.Account] = i
		cs.nonceLocks.Unlock()
	}
}

// NewCoinsBalances is used only to initialize the coin state when chain is not synced.
func NewCoinsBalances() CoinsState {
	return CoinsState{nonceIndex: make(map[[20]byte]int), balanceIndex: make(map[[20]byte]int), Balances: []AccountInfo{}, Nonces: []AccountInfo{}}
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
	i, ok := cs.getBalanceIndex(acc)
	if !ok {
		// Create new nonce and append to the index
	}
	return cs.Nonces[i].Info
}

// SetNonce sets the nonce to a new value.
func (cs *CoinsState) SetNonce(acc [20]byte, value uint64) {
	i, ok := cs.getNonceIndex(acc)
	if !ok {
		// Reduce the balance of a non-existing account (this should never happen)
		return
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
	i, ok := cs.getBalanceIndex(acc)
	if !ok {
		// Append to the slice and add to the map
	}
	cs.Balances[i].Info += amount
	return
}

func (cs *CoinsState) getBalanceIndex(acc [20]byte) (int, bool) {
	cs.balanceLock.Lock()
	defer cs.balanceLock.Unlock()
	balanceIndex, ok := cs.balanceIndex[acc]
	return balanceIndex, ok
}

func (cs *CoinsState) setBalanceIndex(acc [20]byte) {
	cs.balanceLock.Lock()
	defer cs.balanceLock.Unlock()
	newIndex := len(cs.Balances)
	cs.balanceIndex[acc] = newIndex
	return
}

func (cs *CoinsState) getNonceIndex(acc [20]byte) (int, bool) {
	cs.nonceLocks.Lock()
	defer cs.nonceLocks.Unlock()
	nonceIndex, ok := cs.nonceIndex[acc]
	return nonceIndex, ok
}
