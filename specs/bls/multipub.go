package bls

type Multipub struct {
	PublicKeys [][]byte `ssz-size:"?,32" ssz-max:"16777216"`
	NumNeeded  uint16
}

type Multisig struct {
	PublicKey  *Multipub
	Signatures [][]byte `ssz-size:"?,32" ssz-max:"32"`
	KeysSigned []byte `ssz-max:"32"`
}