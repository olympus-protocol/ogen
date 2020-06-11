package bls

import (
	"fmt"
	"io"

	"github.com/olympus-protocol/bls-go/bls"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/pkg/errors"
)

// PublicKey used in the BLS signature scheme.
type PublicKey struct {
	p *bls.PublicKey
}

// Encode encodes to the given writer.
func (p *PublicKey) Encode(w io.Writer) error {
	pubBytes := p.Marshal()
	_, err := w.Write(pubBytes)

	return err
}

// Decode decodes from the given reader.
func (p *PublicKey) Decode(r io.Reader) error {
	pubBytes := make([]byte, 48)
	if _, err := io.ReadFull(r, pubBytes); err != nil {
		return err
	}

	return p.p.Deserialize(pubBytes)
}

// PublicKeyFromBytes creates a BLS public key from a  BigEndian byte slice.
func PublicKeyFromBytes(pub []byte) (*PublicKey, error) {
	if len(pub) != 48 {
		return nil, fmt.Errorf("public key must be %d bytes", 48)
	}
	if cv, ok := pubkeyCache.Get(string(pub)); ok {
		return cv.(*PublicKey).Copy(), nil
	}
	pubKey := &bls.PublicKey{}
	err := pubKey.Deserialize(pub)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal bytes into public key")
	}
	pubKeyObj := &PublicKey{p: pubKey}
	copiedKey := pubKeyObj.Copy()
	pubkeyCache.Set(string(pub), copiedKey, 48)
	return pubKeyObj, nil
}

// Marshal a public key into a LittleEndian byte slice.
func (p *PublicKey) Marshal() []byte {
	return p.p.Serialize()
}

// Equals checks if two public keys are equal.
func (p *PublicKey) Equals(other *PublicKey) bool {
	return p.p.IsEqual(other.p)
}

// Copy the public key to a new pointer reference.
func (p *PublicKey) Copy() *PublicKey {
	np := *p.p
	return &PublicKey{p: &np}
}

// Aggregate two public keys.
func (p *PublicKey) Aggregate(p2 *PublicKey) *PublicKey {
	p.p.Add(p2.p)
	return p
}

// ToAddress converts the public key to a Bech32 address.
func (p *PublicKey) ToAddress(pubPrefix string) (string, error) {
	out := make([]byte, 20)
	pkS := p.p.Serialize()
	h := chainhash.HashH(pkS[:])
	copy(out[:], h[:20])
	return bech32.Encode(pubPrefix, out), nil
}

// Hash calculates the hash of the public key.
func (p *PublicKey) Hash() [20]byte {
	pkS := p.p.Serialize()
	h := chainhash.HashH(pkS[:])
	var hBytes [20]byte
	copy(hBytes[:], h[:])
	return hBytes
}

// Type returns the type of the public key.
func (p *PublicKey) Type() FunctionalSignatureType {
	return TypeSingle
}

var _ FunctionalPublicKey = &PublicKey{}
