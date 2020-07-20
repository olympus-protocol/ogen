package primitives

import (
	fastssz "github.com/ferranbt/fastssz"
	"github.com/golang/snappy"
	"github.com/prysmaticlabs/go-ssz"
)

// AccountInfo is the information contained into both slices. It represents the account hash and a value.
type AccountInfo struct {
	Account [20]byte
	Info    uint64
}

// CoinsStateSerializable is a struct to properly serialize the coinstate efficiently
type CoinsStateSerializable struct {
	Balances []AccountInfo
	Nonces   []AccountInfo
}

// CoinsState is the state that we use to store accounts balances and Nonces
type CoinsState struct {
	Balances map[[20]byte]uint64
	Nonces   map[[20]byte]uint64
	fastssz.Marshaler
	fastssz.Unmarshaler
}

// Marshal serialize to bytes the struct
func (u *CoinsState) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(u)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal deserialize the bytes to struct
func (u *CoinsState) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(d, u)
}

// MarshalSSZ uses the fastssz interface to override the ssz Marshal function
func (u *CoinsState) MarshalSSZ() ([]byte, error) {
	b := []byte{}
	return u.MarshalSSZTo(b)
}

// MarshalSSZTo utility function to match the fastssz interface
func (u *CoinsState) MarshalSSZTo(dst []byte) ([]byte, error) {
	state := u.ToSerializable()
	mb, err := ssz.Marshal(state)
	if err != nil {
		return nil, err
	}
	copy(dst, mb)
	return mb, nil
}

// SizeSSZ returns the Size of the struct
func (u *CoinsState) SizeSSZ() int {
	size := 0
	size += len(u.Balances) * 28
	size += len(u.Nonces) * 28
	return size
}

// UnmarshalSSZ overrides the ssz unmarshal function using a different struct
func (u *CoinsState) UnmarshalSSZ(b []byte) error {
	serializable := new(CoinsStateSerializable)
	err := ssz.Unmarshal(b, serializable)
	if err != nil {
		return err
	}
	u.FromSerializable(serializable)
	return nil
}

// GetTotal returns the total supply on the state
func (u *CoinsState) GetTotal() uint64 {
	total := uint64(0)
	for _, a := range u.Balances {
		total += a
	}
	return total
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

// FromSerializable creates a CoinsState
func (u *CoinsState) FromSerializable(ser *CoinsStateSerializable) {
	u.Balances = map[[20]byte]uint64{}
	u.Nonces = map[[20]byte]uint64{}
	for _, b := range ser.Balances {
		u.Balances[b.Account] = b.Info
	}
	for _, n := range ser.Nonces {
		u.Nonces[n.Account] = n.Info
	}
	return
}

// ToSerializable converts the struct from maps to slices.
func (u *CoinsState) ToSerializable() CoinsStateSerializable {
	balances := []AccountInfo{}
	nonces := []AccountInfo{}
	for k, v := range u.Balances {
		balances = append(balances, AccountInfo{Account: k, Info: v})
	}
	for k, v := range u.Nonces {
		nonces = append(nonces, AccountInfo{Account: k, Info: v})
	}
	return CoinsStateSerializable{Balances: balances, Nonces: nonces}
}
