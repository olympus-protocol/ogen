package bls_blst

import (
	"errors"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	blst "github.com/supranational/blst/bindings/go"
)

// bls12SecretKey used in the BLS signature scheme.
type bls12SecretKey struct {
	p *blst.SecretKey
}

// RandKey creates a new private key using a random method provided as an io.Reader.
func (b BlstImplementation) RandKey() bls_interface.SecretKey {
	// Generate 32 bytes of randomness
	var ikm [32]byte
	_, err := bls_interface.NewGenerator().Read(ikm[:])
	if err != nil {
		return nil
	}
	return &bls12SecretKey{blst.KeyGen(ikm[:])}
}

// SecretKeyFromBytes creates a BLS private key from a BigEndian byte slice.
func (b BlstImplementation) SecretKeyFromBytes(privKey []byte) (bls_interface.SecretKey, error) {
	secKey := new(blst.SecretKey).Deserialize(privKey)
	if secKey == nil {
		return nil, errors.New("could not unmarshal bytes into secret key")
	}

	return &bls12SecretKey{p: secKey}, nil
}

// PublicKey obtains the public key corresponding to the BLS secret key.
func (s *bls12SecretKey) PublicKey() bls_interface.PublicKey {
	return &PublicKey{p: new(blstPublicKey).From(s.p)}
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
	signature := new(blstSignature).Sign(s.p, msg, dst)
	return &Signature{s: signature}
}

// Marshal a secret key into a LittleEndian byte slice.
func (s *bls12SecretKey) Marshal() []byte {
	keyBytes := s.p.Serialize()
	return keyBytes
}

// ToWIF converts the private key to a Bech32 encoded string.
func (s *bls12SecretKey) ToWIF() string {
	return bech32.Encode(bls_interface.Prefix.Private, s.Marshal())
}
