package bls

import (
	bls12 "github.com/herumi/bls-eth-go-binary/bls"
	"github.com/olympus-protocol/ogen/pkg/bech32"
)

// SecretKey used in the BLS signature scheme.
type SecretKey struct {
	p *bls12.SecretKey
}

// RandKey creates a new private key using a random method provided as an io.Reader.
func RandKey() (*SecretKey, error) {
	secKey := &bls12.SecretKey{}
	secKey.SetByCSPRNG()
	if secKey.IsZero() {
		return nil, ErrZeroSecKey
	}
	return &SecretKey{secKey}, nil
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
	return s.p.Serialize()
}

// ToWIF converts the private key to a Bech32 encoded string.
func (s *SecretKey) ToWIF() string {
	return bech32.Encode(Prefix.Private, s.Marshal())
}
