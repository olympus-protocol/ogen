package bls

import (
	"fmt"
)

// CombinedSignature is a signature and a public key meant to match
// the same interface as Multisig.
type CombinedSignature struct {
	sig []byte `ssz-size:"96"`
	pub []byte `ssz-size:"48"`
}

// NewCombinedSignature creates a new combined signature
func NewCombinedSignature(pub *PublicKey, sig *Signature) *CombinedSignature {
	return &CombinedSignature{
		pub: pub,
		sig: sig,
	}
}

// ToSig outputs the bundled signature.
func (cs *CombinedSignature) ToSig() Signature {
	return cs.sig
}

// ToPub outputs the bundled public key.
func (cs *CombinedSignature) ToPub() PublicKey {
	return cs.pub
}

// Sign signs a message using the secret key.
func (cs *CombinedSignature) Sign(sk *SecretKey, msg []byte) error {
	expectedPub := sk.PublicKey()
	if !expectedPub.Equals(&cs.pub) {
		return fmt.Errorf("expected key for %x, but got %x", cs.pub.Marshal(), expectedPub.Marshal())
	}

	sig := sk.Sign(msg)
	cs.sig = *sig
	return nil
}

// Verify verified a message using the secret key.
func (cs *CombinedSignature) Verify(msg []byte) bool {
	return cs.sig.Verify(msg, &cs.pub)
}

// Copy copies the combined signature.
func (cs *CombinedSignature) Copy() *CombinedSignature {
	newCs := &CombinedSignature{}
	s := cs.sig.Copy()
	p := cs.pub.Copy()

	newCs.sig = *s
	newCs.pub = *p

	return newCs
}
