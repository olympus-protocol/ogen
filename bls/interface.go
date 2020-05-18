// Copyright (c) 2019 Phore Project

package bls

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	bls "github.com/phoreproject/bls/g1pubs"
)

type Prefixes struct {
	PubKey          string
	PrivKey         string
	ContractPubKey  string
	ContractPrivKey string
}

// Signature used in the BLS signature scheme.
type Signature struct {
	s bls.Signature
}

// Serialize gets the binary representation of the signature.
func (s Signature) Serialize() [96]byte {
	return s.s.Serialize()
}

// Copy returns a copy of the signature.
func (s Signature) Copy() *Signature {
	c := s.s.Copy()
	return &Signature{*c}
}

// DeserializeSignature deserializes a binary signature into the actual signature.
func DeserializeSignature(b [96]byte) (*Signature, error) {
	s, err := bls.DeserializeSignature(b)
	if err != nil {
		return nil, err
	}

	return &Signature{s: *s}, nil
}

// SecretKey used in the BLS scheme.
type SecretKey struct {
	s bls.SecretKey
}

func NewSecret(s bls.SecretKey) *SecretKey {
	return &SecretKey{s: s}
}

// RandSecretKey generates a random key given a byte reader.
func RandSecretKey(r io.Reader) (*SecretKey, error) {
	key, err := bls.RandKey(r)
	if err != nil {
		return nil, err
	}
	return &SecretKey{s: *key}, nil
}

// DerivePublicKey derives a public key from a secret key.
func (s SecretKey) DerivePublicKey() *PublicKey {
	pub := bls.PrivToPub(&s.s)
	return &PublicKey{p: *pub}
}

// Serialize serializes a secret key to bytes.
func (s SecretKey) Serialize() [32]byte {
	return s.s.Serialize()
}

func (s SecretKey) ToBech32(prefixes Prefixes, contract bool) string {
	ser := s.Serialize()
	buf := bytes.NewBuffer([]byte{})
	buf.Write(ser[:])
	var prefix string
	if contract {
		prefix = prefixes.ContractPrivKey
	} else {
		prefix = prefixes.PrivKey
	}
	return bech32.Encode(prefix, buf.Bytes())
}

// DeserializeSecretKey deserializes a secret key from bytes.
func DeserializeSecretKey(b [32]byte) (*SecretKey, error) {
	k := bls.DeserializeSecretKey(b)
	if k == nil {
		return nil, fmt.Errorf("invalid secret key")
	}
	return &SecretKey{*k}, nil
}

// DeriveSecretKey deserializes a secret key from bytes.
func DeriveSecretKey(b [32]byte) SecretKey {
	k := bls.DeriveSecretKey(b)
	return SecretKey{*k}
}

func NewSecretFromBech32(secret string, prefixes Prefixes, contract bool) (*SecretKey, error) {
	var prefix string
	if contract {
		prefix = prefixes.ContractPrivKey
	} else {
		prefix = prefixes.PrivKey
	}
	net, privKeyBytes, err := bech32.Decode(secret)
	if err != nil {
		return nil, err
	}
	if net != prefix {
		return nil, errors.New("key networks doesn't match")
	}
	var rawPriv [32]byte
	buf := bytes.NewBuffer(rawPriv[:0])
	buf.Write(privKeyBytes)
	return DeserializeSecretKey(rawPriv)
}

// PublicKey corresponding to secret key used in the BLS scheme.
type PublicKey struct {
	p bls.PublicKey
}

// NewPublicKey constructs a new public key based on a raw public key from
// the BLS library.
func NewPublicKey(p *bls.PublicKey) *PublicKey {
	return &PublicKey{p: *p}
}

// ToBech32 converts the public key to a Bech32 address.
func (p PublicKey) ToBech32(prefixes Prefixes) (string, error) {
	out := make([]byte, 20)
	pkS := p.p.Serialize()
	h := chainhash.HashH(pkS[:])
	copy(out[:], h[:20])
	return bech32.Encode(prefixes.PubKey, out), nil
}

// Hash calculates the hash of the public key.
func (p PublicKey) Hash() [20]byte {
	pkS := p.p.Serialize()
	h := chainhash.HashH(pkS[:])
	var hBytes [20]byte
	copy(hBytes[:], h[:])
	return hBytes
}

func (p PublicKey) String() string {
	return p.p.String()
}

// Serialize serializes a public key to bytes.
func (p PublicKey) Serialize() [48]byte {
	return p.p.Serialize()
}

// Equals checks if two public keys are equal.
func (p PublicKey) Equals(other PublicKey) bool {
	return p.p.Equals(other.p)
}

// DeserializePublicKey deserialies a public key from the provided bytes.
func DeserializePublicKey(b [48]byte) (*PublicKey, error) {
	p, err := bls.DeserializePublicKey(b)
	if err != nil {
		return nil, err
	}
	return &PublicKey{*p}, nil
}

// Copy returns a copy of the public key
func (p PublicKey) Copy() PublicKey {
	return p
}

// Sign a message using a secret key - in a beacon/validator client, this key will come from and be unlocked from the account keystore.
func Sign(sec *SecretKey, msg []byte) (*Signature, error) {
	s := bls.Sign(msg, &sec.s)
	return &Signature{s: *s}, nil
}

// VerifySig against a public key.
func VerifySig(pub *PublicKey, msg []byte, sig *Signature) (bool, error) {
	return bls.Verify(msg, &pub.p, &sig.s), nil
}

// AggregateSigs puts multiple signatures into one using the underlying BLS sum functions.
func AggregateSigs(sigs []*Signature) (*Signature, error) {
	blsSigs := make([]*bls.Signature, len(sigs))
	for i := range sigs {
		blsSigs[i] = &sigs[i].s
	}
	aggSig := bls.AggregateSignatures(blsSigs)
	return &Signature{s: *aggSig}, nil
}

// VerifyAggregate verifies a signature over many messages.
func VerifyAggregate(pubkeys []*PublicKey, msgs [][]byte, signature *Signature) bool {
	if len(pubkeys) != len(msgs) {
		return false
	}

	blsPubs := make([]*bls.PublicKey, len(pubkeys))
	for i := range pubkeys {
		blsPubs[i] = &pubkeys[i].p
	}

	return signature.s.VerifyAggregate(blsPubs, msgs)
}

// VerifyAggregateCommon verifies a signature over a common message.
func VerifyAggregateCommon(pubkeys []*PublicKey, msg []byte, signature *Signature) bool {
	blsPubs := make([]*bls.PublicKey, len(pubkeys))
	for i := range pubkeys {
		blsPubs[i] = &pubkeys[i].p
	}

	return signature.s.VerifyAggregateCommon(blsPubs, msg)
}

// AggregatePubKeys aggregates some public keys into one.
func AggregatePubKeys(pubkeys []*PublicKey) *PublicKey {
	blsPubs := make([]*bls.PublicKey, len(pubkeys))
	for i := range pubkeys {
		blsPubs[i] = &pubkeys[i].p
	}

	newPub := bls.AggregatePublicKeys(blsPubs)

	return &PublicKey{p: *newPub}
}

// AggregatePubKey adds another public key to this one.
func (p *PublicKey) AggregatePubKey(other *PublicKey) {
	p.p.Aggregate(&other.p)
}

// AggregateSig adds another signature to this one.
func (s *Signature) AggregateSig(other *Signature) {
	s.s.Aggregate(&other.s)
}

// NewAggregatePublicKey creates a blank public key.
func NewAggregatePublicKey() *PublicKey {
	pub := bls.NewAggregatePubkey()
	return &PublicKey{p: *pub}
}

// NewAggregateSignature creates a blank signature key.
func NewAggregateSignature() *Signature {
	sig := bls.NewAggregateSignature()
	return &Signature{s: *sig}
}
