package bls_kilic

import (
	"crypto/rand"
	bls12 "github.com/kilic/bls12-381"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"math/big"
)

// bls12SecretKey used in the BLS signature scheme.
type bls12SecretKey struct {
	p *big.Int
}

// RandKey creates a new private key using a random method provided as an io.Reader.
func (b KilicImplementation) RandKey() bls_interface.SecretKey {
	k, _ := rand.Int(rand.Reader, bls_interface.RFieldModulus)
	return &bls12SecretKey{p: k}
}

// SecretKeyFromBytes creates a BLS private key from a BigEndian byte slice.
func (b KilicImplementation) SecretKeyFromBytes(privKey []byte) (bls_interface.SecretKey, error) {
	// TODO make sure point is valid on curve
	p := new(big.Int)
	p.SetBytes(privKey)
	return &bls12SecretKey{p: p}, nil
}

// PublicKey obtains the public key corresponding to the BLS secret key.
func (s *bls12SecretKey) PublicKey() bls_interface.PublicKey {
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
func (s *bls12SecretKey) Sign(msg []byte) bls_interface.Signature {
	g2 := bls12.NewG2()
	p, _ := g2.HashToCurve(msg, dst)
	g2.MulScalar(p, p, s.p)
	return &Signature{s: p}
}

// Marshal a secret key into a LittleEndian byte slice.
func (s *bls12SecretKey) Marshal() []byte {
	return s.p.Bytes()
}

// ToWIF converts the private key to a Bech32 encoded string.
func (s *bls12SecretKey) ToWIF() string {
	return bech32.Encode(bls_interface.Prefix.Private, s.Marshal())
}
