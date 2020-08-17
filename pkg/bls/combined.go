package bls

import (
	"errors"
	"fmt"
)

var (
	// ErrorCombinedSignatureSize returns when serialized CombinedSignature size exceed MaxCombinedSignatureSize.
	ErrorCombinedSignatureSize = errors.New("combined signature too big")
)

const (
	// MaxCombinedSignatureSize is the maximum amount of bytes a CombinedSignature can contain.
	MaxCombinedSignatureSize = 96 + 48
)

// CombinedSignature is a signature and a public key meant to match the same interface as Multisig.
type CombinedSignature struct {
	S [96]byte
	P [48]byte
}

// Marshal encodes the data.
func (c *CombinedSignature) Marshal() ([]byte, error) {
	b, err := c.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxCombinedSignatureSize {
		return nil, ErrorCombinedSignatureSize
	}
	return b, nil
}

// Unmarshal decodes the data.
func (c *CombinedSignature) Unmarshal(b []byte) error {
	if len(b) > MaxCombinedSignatureSize {
		return ErrorCombinedSignatureSize
	}
	return c.UnmarshalSSZ(b)
}

// NewCombinedSignature creates a new combined signature
func NewCombinedSignature(pub *PublicKey, sig *Signature) *CombinedSignature {
	var s [96]byte
	var p [48]byte
	copy(s[:], sig.Marshal())
	pubB := pub.Marshal()
	copy(p[:], pubB)
	return &CombinedSignature{
		P: p,
		S: s,
	}
}

// Sig outputs the bundled signature.
func (c *CombinedSignature) Sig() (*Signature, error) {
	return SignatureFromBytes(c.S)
}

// Pub outputs the bundled public key.
func (c *CombinedSignature) Pub() (*PublicKey, error) {
	return PublicKeyFromBytes(c.P)
}

// Sign signs a message using the secret key.
func (c *CombinedSignature) Sign(sk *SecretKey, msg []byte) error {
	expectedPub := sk.PublicKey()
	pub, err := c.Pub()
	if err != nil {
		return err
	}
	if !expectedPub.Equals(pub) {
		return fmt.Errorf("expected key for %x, but got %x", c.P, expectedPub.Marshal())
	}
	var s [96]byte
	sig := sk.Sign(msg)
	copy(s[:], sig.Marshal())
	c.S = s
	return nil
}

// Verify verified a message using the secret key.
func (c *CombinedSignature) Verify(msg []byte) bool {
	sig, err := c.Sig()
	if err != nil {
		return false
	}
	pub, err := c.Pub()
	if err != nil {
		return false
	}
	return sig.Verify(msg, pub)
}

// Copy copies the combined signature.
func (c *CombinedSignature) Copy() *CombinedSignature {
	newCs := new(CombinedSignature)
	copy(newCs.P[:], c.P[:])
	copy(newCs.S[:], c.S[:])
	return newCs
}
