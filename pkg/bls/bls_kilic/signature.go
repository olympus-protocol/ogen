package bls_kilic

import (
	bls12 "github.com/kilic/bls12-381"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
)

var dst = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

// Signature used in the BLS signature scheme.
type Signature struct {
	s *bls12.PointG2
}

// Copy returns a full deep copy of a signature.
func (s *Signature) Copy() bls_interface.Signature {
	ng2 := bls12.PointG2{}
	ng2.Set(s.s)
	return &Signature{s: &ng2}
}

// Marshal a signature into a LittleEndian byte slice.
func (s *Signature) Marshal() []byte {
	return bls12.NewG2().ToCompressed(s.s)
}

// Verify a bls signature given a public key amd a msg
func (s *Signature) Verify(pub bls_interface.PublicKey, msg []byte) bool {
	return false
}

// AggregateVerify verifies each public key against its respective message.
// This is vulnerable to rogue public-key attack. Each user must
// provide a proof-of-knowledge of the public key.
func (s *Signature) AggregateVerify(pubKeys []bls_interface.PublicKey, msg [][32]byte) bool {
	return false
}

// FastAggregateVerify verifies each public key against its respective message.
// This is vulnerable to rogue public-key attack. Each user must
// provide a proof-of-knowledge of the public key.
func (s *Signature) FastAggregateVerify(pubKeys []bls_interface.PublicKey, msg [32]byte) bool {
	return false
}
