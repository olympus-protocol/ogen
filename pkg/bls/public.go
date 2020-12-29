package bls

import (
	bls12381 "github.com/kilic/bls12-381"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// PublicKey used in the BLS signature scheme.
type PublicKey struct {
	p *bls12381.PointG1
}

// Marshal a public key into a LittleEndian byte slice.
func (p *PublicKey) Marshal() []byte {
	return bls12381.NewG1().ToCompressed(p.p)
}

// Copy the public key to a new pointer reference.
func (p *PublicKey) Copy() *PublicKey {
	np := *p.p
	return &PublicKey{p: &np}
}

// Aggregate two public keys.
func (p *PublicKey) Aggregate(p2 *PublicKey) *PublicKey {
	return &PublicKey{p: bls12381.NewG1().Add(p.p, p.p, p2.p)}
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
func (p *PublicKey) ToAccount() string {
	hash, _ := p.Hash()
	return bech32.Encode(Prefix.Public, hash[:])
}
