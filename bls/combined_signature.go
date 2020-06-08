package bls

import (
	"fmt"
	"io"
)

// CombinedSignature is a signature and a public key meant to match
// the same interface as Multisig.
type CombinedSignature struct {
	sig Signature
	pub PublicKey
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

// Encode encodes the combined signature to the writer.
func (cs *CombinedSignature) Encode(w io.Writer) error {
	if err := cs.pub.Encode(w); err != nil {
		return err
	}
	if err := cs.sig.Encode(w); err != nil {
		return err
	}

	return nil
}

// Decode decodes the combined signature from the reader.
func (cs *CombinedSignature) Decode(r io.Reader) error {
	if err := cs.pub.Decode(r); err != nil {
		return err
	}
	if err := cs.sig.Decode(r); err != nil {
		return err
	}

	return nil
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

var _ FunctionalSignature = &CombinedSignature{}
