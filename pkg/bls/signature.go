package bls

import (
	bls12 "github.com/kilic/bls12-381"
)

var dst = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

// Signature used in the BLS signature scheme.
type Signature struct {
	s *bls12.PointG2
}

// Verify a bls signature given a public key, a message.
func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	e := bls12.NewEngine()
	// TODO handle error
	ps, _ := bls12.NewG2().HashToCurve(msg, dst)
	g1Neg := &bls12.PointG1{}
	bls12.NewG1().Neg(g1Neg, &bls12.G1One)
	e.AddPair(g1Neg, s.s)
	e.AddPair(pubKey.p, ps)
	return e.Check()
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
	e := bls12.NewEngine()
	g1Neg := &bls12.PointG1{}
	bls12.NewG1().Neg(g1Neg, &bls12.G1One)
	e.AddPair(g1Neg, s.s)
	for i := 0; i < size; i++ {
		g2 := bls12.NewG2()
		// TODO handle error
		ps, _ := g2.HashToCurve(msgs[i][:], dst)
		e.AddPair(pubKeys[i].p, ps)
	}
	return e.Check()
}

// FastAggregateVerify verifies all the provided public keys with their aggregated signature.
func (s *Signature) FastAggregateVerify(pubKeys []*PublicKey, msg [32]byte) bool {
	size := len(pubKeys)
	if size == 0 {
		return false
	}
	e := bls12.NewEngine()
	g1Neg := &bls12.PointG1{}
	// TODO handle error
	ps, _ := bls12.NewG2().HashToCurve(msg[:], dst)
	bls12.NewG1().Neg(g1Neg, &bls12.G1One)
	e.AddPair(g1Neg, s.s)
	for i := 0; i < size; i++ {
		e.AddPair(pubKeys[i].p, ps)
	}
	return e.Check()
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
