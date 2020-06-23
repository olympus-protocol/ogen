package state

type Exit struct {
	ValidatorPubkey []byte `ssz-size:"48"`
	WithdrawPubkey  []byte `ssz-size:"48"`
	Signature       []byte `ssz-size:"96"`
}
