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

// SignatureFromBytes creates a BLS signature from a LittleEndian byte slice.
func (b KilicImplementation) SignatureFromBytes(sig []byte) (bls_interface.Signature, error) {
	p, err := bls12.NewG2().FromCompressed(sig)
	if err != nil {
		return nil, err
	}
	return &Signature{s: p}, nil
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
func (s *Signature) Verify(pubKey bls_interface.PublicKey, msg []byte) bool {
	e := bls12.NewEngine()
	// TODO handle error
	ps, _ := bls12.NewG2().HashToCurve(msg, dst)
	g1Neg := &bls12.PointG1{}
	bls12.NewG1().Neg(g1Neg, &bls12.G1One)
	e.AddPair(g1Neg, s.s)
	e.AddPair(pubKey.(*PublicKey).p, ps)
	return e.Check()
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
func (s *Signature) AggregateVerify(pubKeys []bls_interface.PublicKey, msgs [][32]byte) bool {
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
		// TODO handle error
		ps, _ := bls12.NewG2().HashToCurve(msgs[i][:], dst)
		e.AddPair(pubKeys[i].(*PublicKey).p, ps)
	}
	return e.Check()
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
func (s *Signature) FastAggregateVerify(pubKeys []bls_interface.PublicKey, msg [32]byte) bool {
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
		e.AddPair(pubKeys[i].(*PublicKey).p, ps)
	}
	return e.Check()
}

// NewAggregateSignature creates a blank aggregate signature.
func (b KilicImplementation) NewAggregateSignature() bls_interface.Signature {
	// TODO handle error
	p, _ := bls12.NewG2().HashToCurve([]byte{'m', 'o', 'c', 'k'}, dst)
	return &Signature{s: p}
}

// AggregateSignatures converts a list of signatures into a single, aggregated sig.
func (b KilicImplementation) AggregateSignatures(sigs []bls_interface.Signature) bls_interface.Signature {
	if len(sigs) == 0 {
		return nil
	}
	g2 := bls12.NewG2()
	sig := &bls12.PointG2{}
	for i := 0; i < len(sigs); i++ {
		g2.Add(sig, sig, sigs[i].(*Signature).s)
	}
	return &Signature{s: sig}
}

// Aggregate is an alias for AggregateSignatures, defined to conform to BLS specification.
//
// In IETF draft BLS specification:
// Aggregate(signature_1, ..., signature_n) -> signature: an
//      aggregation algorithm that compresses a collection of signatures
//      into a single signature.
//
// In ETH2.0 specification:
// def Aggregate(signatures: Sequence[BLSSignature]) -> BLSSignature
//
// Deprecated: Use AggregateSignatures.
func (b KilicImplementation) Aggregate(sigs []bls_interface.Signature) bls_interface.Signature {
	return b.AggregateSignatures(sigs)
}

// Marshal a signature into a LittleEndian byte slice.
func (s *Signature) Marshal() []byte {
	return bls12.NewG2().ToCompressed(s.s)
}

// Copy returns a full deep copy of a signature.
func (s *Signature) Copy() bls_interface.Signature {
	p := &bls12.PointG2{}
	p.Set(s.s)
	return &Signature{s: p}
}
