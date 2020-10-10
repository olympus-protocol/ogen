package primitives

import "github.com/golang/snappy"

// AccountInfo is the information contained into both slices. It represents the account hash and a value.
type AccountInfo struct {
	Account [20]byte
	Info    uint64
}

// CoinsStateSerializable is a struct to properly serialize the coinstate efficiently
type CoinsStateSerializable struct {
	Balances []*AccountInfo `ssz-max:"2097152"`
	Nonces   []*AccountInfo `ssz-max:"2097152"`
}

func (c *CoinsStateSerializable) Marshal() ([]byte, error) {
	b, err := c.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

func (c *CoinsStateSerializable) Unmarshal(b []byte) error {
	des, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return c.UnmarshalSSZ(des)
}

// CoinsState is the state that we use to store accounts balances and Nonces
type CoinsState struct {
	Balances map[[20]byte]uint64
	Nonces   map[[20]byte]uint64
}

// Marshal serialize to bytes the struct
func (u *CoinsState) Marshal() ([]byte, error) {
	s := u.ToSerializable()
	b, err := s.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Unmarshal deserialize the bytes to struct
func (u *CoinsState) Unmarshal(b []byte) error {
	us := new(CoinsStateSerializable)
	err := us.Unmarshal(b)
	if err != nil {
		return err
	}
	u.FromSerializable(us)
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
	var balances []*AccountInfo
	var nonces []*AccountInfo
	for k, v := range u.Balances {
		balances = append(balances, &AccountInfo{Account: k, Info: v})
	}
	for k, v := range u.Nonces {
		nonces = append(nonces, &AccountInfo{Account: k, Info: v})
	}
	return CoinsStateSerializable{Balances: balances, Nonces: nonces}
}
