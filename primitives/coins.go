package primitives

// AccountInfo is the information contained into both slices. It represents the account hash and a value.
type AccountInfo struct {
	Account [20]byte `ssz-size:"20"`
	Info    uint64
}

// CoinsStateSerializable is a struct to properly serialize the coinstate efficiently
type CoinsStateSerializable struct {
	Balances []*AccountInfo `ssz-max:"310995116277762"`
	Nonces   []*AccountInfo `ssz-max:"310995116277762"`
}
