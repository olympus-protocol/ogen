package bls

import (
	bls12 "github.com/kilic/bls12-381"
)

var dst = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

// Signature used in the BLS signature scheme.
type Signature struct {
	s *bls12.PointG2
}

// Marshal a signature into a LittleEndian byte slice.
func (s *Signature) Marshal() []byte {
	return bls12.NewG2().ToCompressed(s.s)
}

// Copy returns a full deep copy of a signature.
func (s *Signature) Copy() *Signature {
	np := *s.s
	return &Signature{s: &np}
}

// Verify a bls signature given a public key, a message.
func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	ps, _ := bls12.NewG2().HashToCurve(msg, dst)
	engine.AddPairInv(&bls12.G1One, s.s)
	engine.AddPair(pubKey.p, ps)
	return engine.Result().IsOne()
}

// AggregateVerify verifies each public key against its respective message.
func (s *Signature) AggregateVerify(pubKeys []*PublicKey, msgs [][32]byte) bool {
	size := len(pubKeys)
	if size == 0 {
		return false
	}
	if size != len(msgs) {
		return false
	}
	engine.AddPairInv(&bls12.G1One, s.s)
	for i := 0; i < size; i++ {
		g2 := bls12.NewG2()
		ps, _ := g2.HashToCurve(msgs[i][:], dst)
		engine.AddPair(pubKeys[i].p, ps)
	}
	return engine.Result().IsOne()
}

// FastAggregateVerify verifies all the provided public keys with their aggregated signature.
func (s *Signature) FastAggregateVerify(pubKeys []*PublicKey, msg [32]byte) bool {
	size := len(pubKeys)
	if size == 0 {
		return false
	}
	ps, _ := bls12.NewG2().HashToCurve(msg[:], dst)
	engine.AddPairInv(&bls12.G1One, s.s)
	for _, pub := range pubKeys {
		engine.AddPair(pub.p, ps)
	}
	return engine.Result().IsOne()
}
