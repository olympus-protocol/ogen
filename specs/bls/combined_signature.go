package bls

type CombinedSignature struct {
	sig []byte `ssz-size:"96"`
	pub []byte `ssz-size:"48"`
}

