package bls

import (
	"fmt"
)

// CombinedSignature is a signature and a public key meant to match
// the same interface as Multisig.
type CombinedSignature struct {
	Sig []byte `ssz-size:"96"`
	Pub []byte `ssz-size:"48"`
}

// NewCombinedSignature creates a new combined signature
func NewCombinedSignature(pub []byte, sig []byte) *CombinedSignature {
	return &CombinedSignature{
		Pub: pub,
		Sig: sig,
	}
}

// ToSig outputs the bundled signature.
func (cs *CombinedSignature) ToSig() *Signature {
	sig, _ := SignatureFromBytes(cs.Sig)
	return sig
}

// ToPub outputs the bundled public key.
func (cs *CombinedSignature) ToPub() *PublicKey {
	pub, _ := PublicKeyFromBytes(cs.Pub)
	return pub
}

// Sign signs a message using the secret key.
func (cs *CombinedSignature) Sign(sk *SecretKey, msg []byte) error {
	expectedPub := sk.PublicKey()
	if !expectedPub.Equals(cs.ToPub()) {
		return fmt.Errorf("expected key for %x, but got %x", cs.Pub, expectedPub.Marshal())
	}

	sig := sk.Sign(msg)
	cs.Sig = sig.Marshal()
	return nil
}

// Verify verified a message using the secret key.
func (cs *CombinedSignature) Verify(msg []byte) bool {
	return cs.ToSig().Verify(msg, cs.ToPub())
}

// Copy copies the combined signature.
func (cs *CombinedSignature) Copy() *CombinedSignature {
	newCs := &CombinedSignature{}
	s := cs.Sig
	p := cs.Pub

	newCs.Sig = s
	newCs.Pub = p

	return newCs
}
