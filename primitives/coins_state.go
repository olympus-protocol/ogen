package primitives

type StateElement struct {
	Account []byte `ssz-size:"20"`
	Value   uint64
}
type CoinsState struct {
	Balances []*StateElement `ssz-max:"16777216"`
	Nonces   []*StateElement `ssz-max:"16777216"`
}
