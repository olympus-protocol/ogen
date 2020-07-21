package bls

import (
	"fmt"

	"github.com/olympus-protocol/bls-go/bls"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/pkg/errors"
)

// SecretKey used in the BLS signature scheme.
type SecretKey struct {
	p *bls.SecretKey
}

// RandKey creates a new private key using a random method provided as an io.Reader.
func RandKey() *SecretKey {
	secKey := &bls.SecretKey{}
	secKey.SetByCSPRNG()
	return &SecretKey{secKey}
}

// DeriveSecretKey returns the derived secret key.
func DeriveSecretKey(bs []byte) *SecretKey {
	return &SecretKey{bls.DeriveSecretKey(bs)}
}

// SecretKeyFromBytes creates a BLS private key from a BigEndian byte slice.
func SecretKeyFromBytes(priv []byte) (*SecretKey, error) {
	if len(priv) != 32 {
		return nil, fmt.Errorf("secret key must be %d bytes", 32)
	}
	secKey := &bls.SecretKey{}
	err := secKey.Deserialize(priv)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal bytes into secret key")
	}
	return &SecretKey{p: secKey}, err
}

// ToWIF converts the private key to a Bech32 encoded string.
func (s *SecretKey) ToWIF(privPrefix string) (string, error) {
	return bech32.Encode(privPrefix, s.Marshal()), nil
}

// PublicKey obtains the public key corresponding to the BLS secret key.
func (s *SecretKey) PublicKey() *PublicKey {
	return &PublicKey{p: s.p.GetPublicKey()}
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
	signature := s.p.SignByte(msg)
	return &Signature{s: signature}
}

// Marshal a secret key into a LittleEndian byte slice.
func (s *SecretKey) Marshal() []byte {
	keyBytes := s.p.Serialize()
	if len(keyBytes) < 32 {
		emptyBytes := make([]byte, 32-len(keyBytes))
		keyBytes = append(emptyBytes, keyBytes...)
	}
	return keyBytes
}

// Add adds two secret keys together.
func (s *SecretKey) Add(other *SecretKey) {
	s.p.Add(other.p)
}
