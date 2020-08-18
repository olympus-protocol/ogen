package bls_herumi

import (
	bls12 "github.com/olympus-protocol/bls-go/bls"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"github.com/pkg/errors"
)

// bls12SecretKey used in the BLS signature scheme.
type bls12SecretKey struct {
	p *bls12.SecretKey
}

var _ bls_interface.SecretKey = &bls12SecretKey{}

// RandKey creates a new private key using a random method provided as an io.Reader.
func (h HerumiImplementation) RandKey() bls_interface.SecretKey {
	secKey := &bls12.SecretKey{}
	secKey.SetByCSPRNG()
	return &bls12SecretKey{secKey}
}

// SecretKeyFromBytes creates a BLS private key from a BigEndian byte slice.
func (h HerumiImplementation) SecretKeyFromBytes(privKey []byte) (bls_interface.SecretKey, error) {
	secKey := &bls12.SecretKey{}
	err := secKey.Deserialize(privKey)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal bytes into secret key")
	}
	return &bls12SecretKey{p: secKey}, err
}

// PublicKey obtains the public key corresponding to the BLS secret key.
func (s *bls12SecretKey) PublicKey() bls_interface.PublicKey {
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
func (s *bls12SecretKey) Sign(msg []byte) bls_interface.Signature {
	signature := s.p.SignByte(msg)
	return &Signature{s: signature}
}

// Marshal a secret key into a LittleEndian byte slice.
func (s *bls12SecretKey) Marshal() []byte {
	keyBytes := s.p.Serialize()
	return keyBytes
}

// ToWIF converts the private key to a Bech32 encoded string.
func (s *bls12SecretKey) ToWIF() (string, error) {
	return bech32.Encode(bls_interface.Prefix.Private, s.Marshal()), nil
}
