package bls

import (
	bls12 "github.com/kilic/bls12-381"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"math/big"
)

var RFieldModulus, _ = new(big.Int).SetString("52435875175126190479447740508185965837690552500527637822603658699938581184513", 10)

// SecretKey used in the BLS signature scheme.
type SecretKey struct {
	p *big.Int
}

// PublicKey obtains the public key corresponding to the BLS secret key.
func (s *SecretKey) PublicKey() *PublicKey {
	p := &bls12.PointG1{}
	return &PublicKey{p: bls12.NewG1().MulScalar(p, &bls12.G1One, s.p)}
}

// Sign a message using a secret key - in a beacon/validator client.
//
// In IETF draft BLS specification:
// Sign(SK, message) -> signature: a signing algorithm that generates
//      a deterministic signature given a secret key SK and a message.
//
// In ETH2.0 specification:
// def Sign(SK: int, message: Bytes) -> BLSSignature
func (s *SecretKey) Sign(msg []byte) *Signature {
	p, _ := bls12.NewG2().HashToCurve(msg, dst)
	bls12.NewG2().MulScalar(p, p, s.p)
	return &Signature{s: p}
}

// Marshal a secret key into a LittleEndian byte slice.
func (s *SecretKey) Marshal() []byte {
	var b [32]byte
	copy(b[:], s.p.Bytes())
	return b[:]
}

// ToWIF converts the private key to a Bech32 encoded string.
func (s *SecretKey) ToWIF() string {
	return bech32.Encode(Prefix.Private, s.Marshal())
}
