package bls

import (
	"fmt"

	"github.com/olympus-protocol/bls-go/bls"
	"github.com/pkg/errors"
)

// Signature used in the BLS signature scheme.
type Signature struct {
	s *bls.Sign
}

// Copy returns a copy of the signature.
func (s *Signature) Copy() *Signature {
	sCopy := *s.s
	return &Signature{
		s: &sCopy,
	}
}

// SignatureFromBytes creates a BLS signature from a LittleEndian byte slice.
func SignatureFromBytes(sig []byte) (*Signature, error) {
	if len(sig) != 96 {
		return nil, fmt.Errorf("signature must be %d bytes", 96)
	}
	signature := &bls.Sign{}
	err := signature.Deserialize(sig)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal bytes into signature")
	}
	return &Signature{s: signature}, nil
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
func (s *Signature) Verify(msg []byte, pub *PublicKey) bool {
	return s.s.VerifyByte(pub.p, msg)
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
	size := len(pubKeys)
	if size == 0 {
		return false
	}
	if size != len(msgs) {
		return false
	}
	msgSlices := []byte{}
	var rawKeys []bls.PublicKey
	for i := 0; i < size; i++ {
		msgSlices = append(msgSlices, msgs[i][:]...)
		rawKeys = append(rawKeys, *pubKeys[i].p)
	}
	return s.s.AggregateVerify(rawKeys, msgSlices)
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
	if len(pubKeys) == 0 {
		return false
	}
	rawKeys := make([]bls.PublicKey, len(pubKeys))
	for i := 0; i < len(pubKeys); i++ {
		rawKeys[i] = *pubKeys[i].p
	}

	return s.s.FastAggregateVerify(rawKeys, msg[:])
}

// NewAggregateSignature creates a blank aggregate signature.
func NewAggregateSignature() *Signature {
	return &Signature{s: bls.HashAndMapToSignature([]byte{'m', 'o', 'c', 'k'})}
}

// AggregateSignatures converts a list of signatures into a single, aggregated sig.
func AggregateSignatures(sigs []*Signature) *Signature {
	if len(sigs) == 0 {
		return nil
	}

	// Copy signature
	signature := *sigs[0].s
	for i := 1; i < len(sigs); i++ {
		signature.Add(sigs[i].s)
	}
	return &Signature{s: &signature}
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
func Aggregate(sigs []*Signature) *Signature {
	return AggregateSignatures(sigs)
}

// Marshal a signature into a LittleEndian byte slice.
func (s *Signature) Marshal() []byte {
	return s.s.Serialize()
}
