// Package common provides the BLS interfaces that are implemented by the various BLS wrappers.
//
// This package should not be used by downstream consumers. These interfaces are re-exporter by
// github.com/prysmaticlabs/prysm/shared/bls. This package exists to prevent an import circular
// dependency.
package common

// SecretKey represents a BLS secret or private key.
type SecretKey interface {
	PublicKey() PublicKey
	Sign(msg []byte) Signature
	Marshal() []byte
	IsZero() bool
}

// PublicKey represents a BLS public key.
type PublicKey interface {
	Marshal() []byte
	Copy() PublicKey
	Aggregate(p2 PublicKey) PublicKey
	Hash() ([20]byte, error)
}

// Signature represents a BLS signature.
type Signature interface {
	Verify(pubKey PublicKey, msg []byte) bool
	AggregateVerify(pubKeys []PublicKey, msgs [][32]byte) bool
	FastAggregateVerify(pubKeys []PublicKey, msg [32]byte) bool
	Marshal() []byte
	Copy() Signature
}

// Implementation represents a BLS signatures implementation
type Implementation interface {
	SecretKeyFromBytes(privKey []byte) (SecretKey, error)
	PublicKeyFromBytes(pubKey []byte) (PublicKey, error)
	SignatureFromBytes(sig []byte) (Signature, error)
	AggregatePublicKeys(pubs [][]byte) (PublicKey, error)
	Aggregate(sigs []Signature) Signature
	AggregateSignatures(sigs []Signature) Signature
	NewAggregateSignature() Signature
	RandKey() (SecretKey, error)
}
