package primitives

import "github.com/olympus-protocol/ogen/csmt"

type CoinsState struct {
	balances *csmt.Tree
	nonces   *csmt.Tree
}

func (cs *CoinsState) GetTotal() uint64 {
	return 0
}

func (cs *CoinsState) Get(acc [20]byte) uint64 {
	return 0
}

func (cs *CoinsState) GetNonce(acc [20]byte) uint64 {
	return 0
}

func (cs *CoinsState) Increase(acc [20]byte, balance uint64) {
	return
}

func (cs *CoinsState) Reduce(acc [20]byte, balance uint64) {
	return
}

func (cs *CoinsState) SetNonce(acc [20]byte, nonce uint64) {
	return
}
