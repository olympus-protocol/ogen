package bls_herumi

import (
	bls12 "github.com/olympus-protocol/bls-go/bls"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"github.com/olympus-protocol/ogen/pkg/chainhash"

	"github.com/pkg/errors"
)

// PublicKey used in the BLS signature scheme.
type PublicKey struct {
	p *bls12.PublicKey
}

var _ bls_interface.PublicKey = &PublicKey{}

// PublicKeyFromBytes creates a BLS public key from a  BigEndian byte slice.
func (h HerumiImplementation) PublicKeyFromBytes(pubKey []byte) (bls_interface.PublicKey, error) {
	if cv, ok := pubkeyCache.Get(string(pubKey)); ok {
		return cv.(*PublicKey).Copy(), nil
	}
	p := &bls12.PublicKey{}
	err := p.Deserialize(pubKey)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal bytes into public key")
	}
	pubKeyObj := &PublicKey{p: p}
	pubkeyCache.Set(string(pubKey), pubKeyObj.Copy(), 48)
	return pubKeyObj, nil
}

// Marshal a public key into a LittleEndian byte slice.
func (p *PublicKey) Marshal() []byte {
	return p.p.Serialize()
}

// Copy the public key to a new pointer reference.
func (p *PublicKey) Copy() bls_interface.PublicKey {
	np := *p.p
	return &PublicKey{p: &np}
}

// Aggregate two public keys.
func (p *PublicKey) Aggregate(p2 bls_interface.PublicKey) bls_interface.PublicKey {
	p.p.Add(p2.(*PublicKey).p)
	return p
}

// Hash calculates the hash of the public key.
func (p *PublicKey) Hash() ([20]byte, error) {
	pkS := p.p.Serialize()
	h := chainhash.HashH(pkS[:])
	var hBytes [20]byte
	copy(hBytes[:], h[:])
	return hBytes, nil
}

// ToAccount converts the public key to a Bech32 address.
func (p *PublicKey) ToAccount() string {
	hash, _ := p.Hash()
	return bech32.Encode(bls_interface.Prefix.Public, hash[:])
}
