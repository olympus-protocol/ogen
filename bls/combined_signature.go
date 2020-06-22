package bls

import (
	"fmt"

	"github.com/prysmaticlabs/go-ssz"
)

// CombinedSignature is a signature and a public key meant to match
// the same interface as Multisig.
type CombinedSignature struct {
	sig Signature
	pub PublicKey
}

// Marshal serializes the struct to bytes
func (cs *CombinedSignature) Marshal() []byte {
	ser, _ := ssz.Marshal(cs)
	return ser
}

// Unmarshal deserializes the struct from bytes
func (cs *CombinedSignature) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, cs)
}

// NewCombinedSignature creates a new combined signature
func NewCombinedSignature(pub *PublicKey, sig *Signature) *CombinedSignature {
	return &CombinedSignature{
		pub: *pub,
		sig: *sig,
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

// GetPublicKey gets the functional public key.
func (cs *CombinedSignature) GetPublicKey() FunctionalPublicKey {
	return &cs.pub
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

// Type returns the signature type.
func (cs *CombinedSignature) Type() FunctionalSignatureType {
	return TypeSingle
}

// Copy copies the combined signature.
func (cs *CombinedSignature) Copy() FunctionalSignature {
	newCs := &CombinedSignature{}
	s := cs.sig.Copy()
	p := cs.pub.Copy()

	newCs.sig = *s
	newCs.pub = *p

	return newCs
}

var _ FunctionalSignature = &CombinedSignature{}
