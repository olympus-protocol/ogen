package primitives

import "github.com/prysmaticlabs/go-ssz"

// CoinsState is the state that we
type CoinsState struct {
	Balances map[[20]byte]uint64
	Nonces   map[[20]byte]uint64
}

// Marshal serializes the struct to bytes
func (u *CoinsState) Marshal() ([]byte, error) {
	return ssz.Marshal(u)
}

// Unmarshal deserializes the struct from bytes
func (u *CoinsState) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, u)
}

// Copy copies CoinsState and returns a new one.
func (u *CoinsState) Copy() CoinsState {
	u2 := *u
	u2.Balances = make(map[[20]byte]uint64)
	u2.Nonces = make(map[[20]byte]uint64)
	for i, c := range u.Balances {
		u2.Balances[i] = c
	}
	for i, c := range u.Nonces {
		u2.Nonces[i] = c
	}
	return u2
}
