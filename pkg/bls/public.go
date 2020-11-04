package bls

import (
	bls12 "github.com/herumi/bls-eth-go-binary/bls"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// PublicKey used in the BLS signature scheme.
type PublicKey struct {
	p *bls12.PublicKey
}

// Marshal a public key into a LittleEndian byte slice.
func (p *PublicKey) Marshal() []byte {
	return p.p.Serialize()
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

// IsInfinite checks if the public key is infinite.
func (p *PublicKey) IsInfinite() bool {
	return p.p.IsZero()
}
