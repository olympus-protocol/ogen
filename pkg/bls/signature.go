package bls

import (
	bls12381 "github.com/kilic/bls12-381"
)

// Signature used in the BLS signature scheme.
type Signature struct {
	s *bls12381.PointG2
}

// Verify a bls signature given a public key, a message.
//
// In IETF draft BLS specification:
// Verify(PK, message, signature) -> VALID or INVALID: a verification
//      algorithm that outputs VALID if signature is a valid signature of
//      message under public key PK, and INVALID otherwise.
//
// In ETH2.0 specification:
// def Verify(PK: BLSPubkey, message: Bytes, signature: BLSSignature) -> bool
func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {

	e := bls12381.NewEngine()

	e.AddPairInv(&bls12381.G1One, s.s)

	ps, err := bls12381.NewG2().HashToCurve(msg, dst)
	if err != nil {
		return false
	}

	e.AddPair(pubKey.p, ps)
	return e.Result().IsOne()
}

// AggregateVerify verifies each public key against its respective message.
// This is vulnerable to rogue public-key attack. Each user must
// provide a proof-of-knowledge of the public key.
//
// In IETF draft BLS specification:
// AggregateVerify((PK_1, message_1), ..., (PK_n, message_n),
//      signature) -> VALID or INVALID: an aggregate verification
//      algorithm that outputs VALID if signature is a valid aggregated
//      signature for a collection of public keys and messages, and
//      outputs INVALID otherwise.
//
// In ETH2.0 specification:
// def AggregateVerify(pairs: Sequence[PK: BLSPubkey, message: Bytes], signature: BLSSignature) -> boo
func (s *Signature) AggregateVerify(pubKeys []*PublicKey, msgs [][32]byte) bool {

	e := bls12381.NewEngine()

	size := len(pubKeys)
	if size == 0 {
		return false
	}

	if size != len(msgs) {
		return false
	}

	e.AddPairInv(&bls12381.G1One, s.s)

	for i := 0; i < size; i++ {

		g2 := bls12381.NewG2()

		ps, err := g2.HashToCurve(msgs[i][:], dst)
		if err != nil {
			return false
		}

		e.AddPair(pubKeys[i].p, ps)
	}

	return e.Result().IsOne()
}

// FastAggregateVerify verifies all the provided public keys with their aggregated signature.
//
// In IETF draft BLS specification:
// FastAggregateVerify(PK_1, ..., PK_n, message, signature) -> VALID
//      or INVALID: a verification algorithm for the aggregate of multiple
//      signatures on the same message.  This function is faster than
//      AggregateVerify.
//
// In ETH2.0 specification:
// def FastAggregateVerify(PKs: Sequence[BLSPubkey], message: Bytes, signature: BLSSignature) -> bool
func (s *Signature) FastAggregateVerify(pubKeys []*PublicKey, msg [32]byte) bool {
	e := bls12381.NewEngine()
	size := len(pubKeys)
	if size == 0 {
		return false
	}

	ps, err := bls12381.NewG2().HashToCurve(msg[:], dst)
	if err != nil {
		return false
	}

	e.AddPairInv(&bls12381.G1One, s.s)
	for _, pub := range pubKeys {
		e.AddPair(pub.p, ps)
	}
	return e.Result().IsOne()
}

// Marshal a signature into a LittleEndian byte slice.
func (s *Signature) Marshal() []byte {
	return bls12381.NewG2().ToCompressed(s.s)
}

// Copy returns a full deep copy of a signature.
func (s *Signature) Copy() *Signature {
	np := *s.s
	return &Signature{s: &np}
}
