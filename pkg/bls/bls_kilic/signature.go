package bls_kilic

import (
	"errors"
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
	e := bls12.NewEngine()
	e.AddPair(
		pub.(*PublicKey).p,
		s.s,
	)
	return e.Result().IsOne()
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

// SignatureFromBytes creates a BLS signature from a LittleEndian byte slice.
func (k KilicImplementation) SignatureFromBytes(sig []byte) (bls_interface.Signature, error) {
	s, err := bls12.NewG2().FromCompressed(sig)
	if err != nil {
		return nil, errors.New("could not unmarshal bytes into signature")
	}
	return &Signature{s: s}, nil
}

// NewAggregateSignature creates a blank aggregate signature.
func (k KilicImplementation) NewAggregateSignature() bls_interface.Signature {
	return &Signature{s: &bls12.PointG2{}}
}

// AggregateSignatures converts a list of signatures into a single, aggregated sig.
func (k KilicImplementation) AggregateSignatures(sigs []bls_interface.Signature) bls_interface.Signature {
	aggregated := k.NewAggregateSignature()
	g2 := bls12.NewG2()
	for i := 0; i < len(sigs); i++ {
		sig := sigs[i]
		if sig == nil {
			continue
		}
		g2.Add(aggregated.(*Signature).s, aggregated.(*Signature).s, sig.(*Signature).s)
	}
	return aggregated
}

func (k KilicImplementation) Aggregate(sigs []bls_interface.Signature) bls_interface.Signature {
	return k.AggregateSignatures(sigs)
}

func (k KilicImplementation) VerifyMultipleSignatures(sigs [][]byte, msgs [][32]byte, pubKeys []bls_interface.PublicKey) (bool, error) {
	return false, nil
}
