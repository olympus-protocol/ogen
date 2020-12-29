package bls

import (
	"github.com/kilic/bls12-381"
	"github.com/olympus-protocol/ogen/pkg/bech32"
)

// SecretKey used in the BLS signature scheme.
type SecretKey struct {
	p *bls12381.Fr
}

// PublicKey obtains the public key corresponding to the BLS secret key.
func (s *SecretKey) PublicKey() *PublicKey {
	p := &bls12381.PointG1{}
	return &PublicKey{p: bls12381.NewG1().MulScalar(p, &bls12381.G1One, s.p)}
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
	p, _ := bls12381.NewG2().HashToCurve(msg, dst)
	bls12381.NewG2().MulScalar(p, p, s.p)
	return &Signature{s: p}
}

// Marshal a secret key into a LittleEndian byte slice.
func (s *SecretKey) Marshal() []byte {
	return s.p.ToBytes()
}

// ToWIF converts the private key to a Bech32 encoded string.
func (s *SecretKey) ToWIF() string {
	return bech32.Encode(Prefix.Private, s.Marshal())
}
