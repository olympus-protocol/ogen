// Package bls implements a go-wrapper around a library implementing the
// the bls-381 curve and signature scheme. This package exposes a public API for
// verifying and aggregating BLS signatures used by Ethereum 2.0.
package bls

import (
	"fmt"
	"math/big"

	"github.com/dgraph-io/ristretto"
	"github.com/olympus-protocol/bls-go/bls"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/pkg/errors"
)

func init() {
	if err := bls.Init(bls.BLS12_381); err != nil {
		panic(err)
	}
	if err := bls.SetETHmode(bls.EthModeDraft07); err != nil {
		panic(err)
	}
}

// DomainByteLength length of domain byte array.
const DomainByteLength = 4

var maxKeys = int64(100000)
var pubkeyCache, _ = ristretto.NewCache(&ristretto.Config{
	NumCounters: maxKeys,
	MaxCost:     1 << 19, // 500 kb is cache max size
	BufferItems: 64,
})

// RFieldModulus for the bls-381 curve.
var RFieldModulus, _ = new(big.Int).SetString("52435875175126190479447740508185965837690552500527637822603658699938581184513", 10)

// Signature used in the BLS signature scheme.
type Signature struct {
	s *bls.Sign
}

// PublicKey used in the BLS signature scheme.
type PublicKey struct {
	p *bls.PublicKey
}

// SecretKey used in the BLS signature scheme.
type SecretKey struct {
	p *bls.SecretKey
}

// RandKey creates a new private key using a random method provided as an io.Reader.
func RandKey() *SecretKey {
	secKey := &bls.SecretKey{}
	secKey.SetByCSPRNG()
	return &SecretKey{secKey}
}

func DeriveSecretKey(bs []byte) *SecretKey {
	return &SecretKey{bls.DeriveSecretKey(bs)}
}

// SecretKeyFromBytes creates a BLS private key from a BigEndian byte slice.
func SecretKeyFromBytes(priv []byte) (*SecretKey, error) {
	if len(priv) != 32 {
		return nil, fmt.Errorf("secret key must be %d bytes", 32)
	}
	secKey := &bls.SecretKey{}
	err := secKey.Deserialize(priv)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal bytes into secret key")
	}
	return &SecretKey{p: secKey}, err
}

// ToWIF converts the private key to a Bech32 encoded string.
func (p *PublicKey) ToWIF(privPrefix string) (string, error) {
	return bech32.Encode(privPrefix, p.Marshal()), nil
}

// PublicKeyFromBytes creates a BLS public key from a  BigEndian byte slice.
func PublicKeyFromBytes(pub []byte) (*PublicKey, error) {
	if len(pub) != 48 {
		return nil, fmt.Errorf("public key must be %d bytes", 48)
	}
	if cv, ok := pubkeyCache.Get(string(pub)); ok {
		return cv.(*PublicKey).Copy()
	}
	pubKey := &bls.PublicKey{}
	err := pubKey.Deserialize(pub)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal bytes into public key")
	}
	pubKeyObj := &PublicKey{p: pubKey}
	copiedKey, err := pubKeyObj.Copy()
	if err != nil {
		return nil, errors.Wrap(err, "could not copy public key")
	}
	pubkeyCache.Set(string(pub), copiedKey, 48)
	return pubKeyObj, nil
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

// PublicKey obtains the public key corresponding to the BLS secret key.
func (s *SecretKey) PublicKey() *PublicKey {
	return &PublicKey{p: s.p.GetPublicKey()}
}

// Sign a message using a secret key - in a beacon/validator client.
//
// In IETF draft BLS specification:
// Sign(SK, message) -> signature: a signing algorithm that generates
//      a deterministic signature given a secret key SK and a message.
//
// In ETH2.0 specification:
// def Sign(SK: int, message: Bytes) -> BLSSignature
func (s *SecretKey) Sign(msg []byte) *Signature {
	signature := s.p.SignByte(msg)
	return &Signature{s: signature}
}

// Marshal a secret key into a LittleEndian byte slice.
func (s *SecretKey) Marshal() []byte {
	keyBytes := s.p.Serialize()
	if len(keyBytes) < 32 {
		emptyBytes := make([]byte, 32-len(keyBytes))
		keyBytes = append(emptyBytes, keyBytes...)
	}
	return keyBytes
}

// Add adds two secret keys together.
func (s *SecretKey) Add(other *SecretKey) {
	s.p.Add(other.p)
}

// Marshal a public key into a LittleEndian byte slice.
func (p *PublicKey) Marshal() []byte {
	return p.p.Serialize()
}

// Copy the public key to a new pointer reference.
func (p *PublicKey) Copy() (*PublicKey, error) {
	np := *p.p
	return &PublicKey{p: &np}, nil
}

// Aggregate two public keys.
func (p *PublicKey) Aggregate(p2 *PublicKey) *PublicKey {
	p.p.Add(p2.p)
	return p
}

// ToAddress converts the public key to a Bech32 address.
func (p *PublicKey) ToAddress(pubPrefix string) (string, error) {
	out := make([]byte, 20)
	pkS := p.p.Serialize()
	h := chainhash.HashH(pkS[:])
	copy(out[:], h[:20])
	return bech32.Encode(pubPrefix, out), nil
}

// Hash calculates the hash of the public key.
func (p *PublicKey) Hash() [20]byte {
	pkS := p.p.Serialize()
	h := chainhash.HashH(pkS[:])
	var hBytes [20]byte
	copy(hBytes[:], h[:])
	return hBytes
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
