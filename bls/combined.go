package bls

import (
	"fmt"

	"github.com/golang/snappy"
	"github.com/prysmaticlabs/go-ssz"
)

// CombinedSignature is a signature and a public key meant to match
// the same interface as Multisig.
type CombinedSignature struct {
	S []byte
	P []byte
}

// Marshal encodes the data.
func (cs *CombinedSignature) Marshal() ([]byte, error) {
	return ssz.Marshal(cs)
}

// Unmarshal decodes the data.
func (cs *CombinedSignature) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(d, cs)
}

// NewCombinedSignature creates a new combined signature
func NewCombinedSignature(pub *PublicKey, sig *Signature) *CombinedSignature {
	return &CombinedSignature{
		P: pub.Marshal(),
		S: sig.Marshal(),
	}
}

// ToSig outputs the bundled signature.
func (cs *CombinedSignature) Sig() (*Signature, error) {
	return SignatureFromBytes(cs.S)
}

// ToPub outputs the bundled public key.
func (cs *CombinedSignature) Pub() (*PublicKey, error) {
	return PublicKeyFromBytes(cs.P)
}

// GetPublicKey gets the functional public key.
func (cs *CombinedSignature) GetPublicKey() (FunctionalPublicKey, error) {
	return PublicKeyFromBytes(cs.P)
}

// Sign signs a message using the secret key.
func (cs *CombinedSignature) Sign(sk *SecretKey, msg []byte) error {
	expectedPub := sk.PublicKey()
	pub, err := cs.Pub()
	if err != nil {
		return err
	}
	if !expectedPub.Equals(pub) {
		return fmt.Errorf("expected key for %x, but got %x", cs.P, expectedPub.Marshal())
	}

	sig := sk.Sign(msg)
	cs.S = sig.Marshal()
	return nil
}

// Verify verified a message using the secret key.
func (cs *CombinedSignature) Verify(msg []byte) bool {
	sig, err := cs.Sig()
	if err != nil {
		return false
	}
	pub, err := cs.Pub()
	if err != nil {
		return false
	}
	return sig.Verify(msg, pub)
}

// Type returns the signature type.
func (cs *CombinedSignature) Type() FunctionalSignatureType {
	return TypeSingle
}

// Copy copies the combined signature.
func (cs *CombinedSignature) Copy() FunctionalSignature {
	newCs := &CombinedSignature{}
	newCs.S = cs.S
	newCs.P = cs.P

	return newCs
}

var _ FunctionalSignature = &CombinedSignature{}
