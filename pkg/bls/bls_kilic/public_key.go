package bls_kilic

import (
	bls12 "github.com/kilic/bls12-381"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/pkg/errors"
)

// PublicKey used in the BLS signature scheme.
type PublicKey struct {
	p *bls12.PointG1
}

// Marshal a public key into a LittleEndian byte slice.
func (p *PublicKey) Marshal() []byte {
	return bls12.NewG1().ToCompressed(p.p)
}

// Copy the public key to a new pointer reference.
func (p *PublicKey) Copy() bls_interface.PublicKey {
	np := &PublicKey{p: new(bls12.PointG1).Set(p.p)}
	return np
}

// Aggregate two public keys.
func (p *PublicKey) Aggregate(p2 bls_interface.PublicKey) bls_interface.PublicKey {
	bls12.NewG1().Add(p.p, p.p, p2.(*PublicKey).p)
	return p
}

// Hash calculates the hash of the public key.
func (p *PublicKey) Hash() ([20]byte, error) {
	pkS := p.Marshal()
	h := chainhash.HashH(pkS[:])
	var hBytes [20]byte
	copy(hBytes[:], h[:])
	return hBytes, nil
}

// ToAccount converts the public key to a Bech32 address.
func (p *PublicKey) ToAccount() (string, error) {
	out := make([]byte, 20)
	h := chainhash.HashH(p.Marshal())
	copy(out[:], h[:20])
	return bech32.Encode(bls_interface.Prefix.Public, out), nil
}

// PublicKeyFromBytes creates a BLS public key from a  LittleEndian byte slice.
func (k KilicImplementation) PublicKeyFromBytes(pub []byte) (bls_interface.PublicKey, error) {
	g1Elems := bls12.NewG1()
	p, err := g1Elems.FromCompressed(pub[:])
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal bytes into public key")
	}
	pubkey := &PublicKey{p: p}
	return pubkey, nil
}
