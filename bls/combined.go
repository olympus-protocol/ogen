package bls

import (
	"fmt"

	"github.com/golang/snappy"
)

// CombinedSignature is a signature and a public key meant to match the same interface as Multisig.
type CombinedSignature struct {
	S [96]byte
	P [48]byte
}

// Marshal encodes the data.
func (cs *CombinedSignature) Marshal() ([]byte, error) {
	b, err := cs.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (cs *CombinedSignature) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return cs.UnmarshalSSZ(d)
}

// NewCombinedSignature creates a new combined signature
func NewCombinedSignature(pub *PublicKey, sig *Signature) *CombinedSignature {
	var s [96]byte
	var p [48]byte
	copy(s[:], sig.Marshal())
	copy(p[:], pub.Marshal())
	return &CombinedSignature{
		P: p,
		S: s,
	}
}

// Sig outputs the bundled signature.
func (cs *CombinedSignature) Sig() (*Signature, error) {
	return SignatureFromBytes(cs.S)
}

// Pub outputs the bundled public key.
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
	var s [96]byte
	sig := sk.Sign(msg)
	copy(s[:], sig.Marshal())
	cs.S = s
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
