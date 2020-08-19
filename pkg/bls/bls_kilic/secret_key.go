package bls_kilic

import (
	"crypto/rand"
	"errors"
	bls12 "github.com/kilic/bls12-381"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"math/big"
)

type bls12SecretKey struct {
	p *big.Int
}

// RandKey creates a new private key using a random method provided as an io.Reader.
func (k KilicImplementation) RandKey() bls_interface.SecretKey {
	s, _ := rand.Int(rand.Reader, bls_interface.RFieldModulus)
	return &bls12SecretKey{s}
}

// PublicKey obtains the public key corresponding to the BLS secret key.
func (s *bls12SecretKey) PublicKey() bls_interface.PublicKey {
	p := &bls12.PointG1{}
	bls12.NewG1().MulScalar(p, &bls12.G1One, s.p)
	return &PublicKey{p: p}
}

// Sign a message using a secret key - in a beacon/validator client.
func (s *bls12SecretKey) Sign(msg []byte) bls_interface.Signature {
	g2 := bls12.NewG2()
	signature, _ := g2.HashToCurve(msg, dst)
	g2.MulScalar(signature, signature, s.p)
	return &Signature{s: signature}
}

// Marshal a secret key into a LittleEndian byte slice.
func (s *bls12SecretKey) Marshal() []byte {
	keyBytes := s.p.Bytes()
	if len(keyBytes) < 32 {
		emptyBytes := make([]byte, 32-len(keyBytes))
		keyBytes = append(emptyBytes, keyBytes...)
	}
	return keyBytes
}

// ToWIF converts the private key to a Bech32 encoded string.
func (s *bls12SecretKey) ToWIF() (string, error) {
	return bech32.Encode(bls_interface.Prefix.Private, s.Marshal()), nil
}

func (k KilicImplementation) SecretKeyFromBytes(priv []byte) (bls_interface.SecretKey, error) {
	sk := new(big.Int).SetBytes(priv[:])
	if bls_interface.RFieldModulus.Cmp(sk) < 0 {
		return nil, errors.New("invalid private key")
	}
	return &bls12SecretKey{p: sk}, nil
}
