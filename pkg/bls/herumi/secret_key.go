package herumi
/*
import (
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/params"

	bls12 "github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
)

// bls12SecretKey used in the BLS signature scheme.
type bls12SecretKey struct {
	p *bls12.SecretKey
}

// RandKey creates a new private key using a random method provided as an io.Reader.
func (h *Herumi) RandKey() (common.SecretKey, error) {
	secKey := &bls12.SecretKey{}
	secKey.SetByCSPRNG()
	if secKey.IsZero() {
		return nil, errors.New("generated a zero secret key")
	}
	return &bls12SecretKey{secKey}, nil
}

// SecretKeyFromBytes creates a BLS private key from a BigEndian byte slice.
func (h *Herumi) SecretKeyFromBytes(privKey []byte) (common.SecretKey, error) {
	if len(privKey) != 32 {
		return nil, fmt.Errorf("secret key must be %d bytes", 32)
	}
	secKey := &bls12.SecretKey{}
	err := secKey.Deserialize(privKey)
	if err != nil {
		return nil, common.ErrSecretUnmarshal
	}
	wrappedKey := &bls12SecretKey{p: secKey}
	if wrappedKey.IsZero() {
		return nil, common.ErrZeroKey
	}
	return wrappedKey, err
}

// PublicKey obtains the public key corresponding to the BLS secret key.
func (s *bls12SecretKey) PublicKey() common.PublicKey {
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
func (s *bls12SecretKey) Sign(msg []byte) common.Signature {
	signature := s.p.SignByte(msg)
	return &Signature{s: signature}
}

// Marshal a secret key into a LittleEndian byte slice.
func (s *bls12SecretKey) Marshal() []byte {
	keyBytes := s.p.Serialize()
	if len(keyBytes) < 32 {
		emptyBytes := make([]byte, 32-len(keyBytes))
		keyBytes = append(emptyBytes, keyBytes...)
	}
	return keyBytes
}

// IsZero checks if the secret key is a zero key.
func (s *bls12SecretKey) IsZero() bool {
	return s.p.IsZero()
}

// ToWIF converts the private key to a Bech32 encoded string.
func (s *bls12SecretKey) ToWIF(p *params.AccountPrefixes) string {
	return bech32.Encode(p.Private, s.Marshal())
}
*/