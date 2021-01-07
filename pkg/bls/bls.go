// Package bls implements a go-wrapper around a library implementing the
// the BLS12-381 curve and signature scheme. This package exposes a public API for
// verifying and aggregating BLS signatures used by Ethereum 2.0.
package bls

import (
	"github.com/olympus-protocol/ogen/pkg/bls/blst"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/bls/herumi"
	"github.com/olympus-protocol/ogen/pkg/params"
	"math/big"

	"github.com/pkg/errors"
)

var currImplementation common.Implementation = herumi.NewHerumiInterface()

// KeyPair is an interface struct to serve keypairs
type KeyPair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

var Prefix = params.MainNet.AccountPrefixes

func Initialize(c *params.ChainParams, impl string) {
	Prefix = c.AccountPrefixes
	switch impl {
	case "herumi":
		currImplementation = herumi.NewHerumiInterface()
	case "blst":
		currImplementation = blst.NewBLSTInterface()
	}
}

// SecretKeyFromBytes creates a BLS private key from a BigEndian byte slice.
func SecretKeyFromBytes(privKey []byte) (SecretKey, error) {
	return currImplementation.SecretKeyFromBytes(privKey)
}

// SecretKeyFromBigNum takes in a big number string and creates a BLS private key.
func SecretKeyFromBigNum(s string) (SecretKey, error) {
	num := new(big.Int)
	num, ok := num.SetString(s, 10)
	if !ok {
		return nil, errors.New("could not set big int from string")
	}
	bts := num.Bytes()
	// Pad key at the start with zero bytes to make it into a 32 byte key.
	if len(bts) < 32 {
		emptyBytes := make([]byte, 32-len(bts))
		bts = append(emptyBytes, bts...)
	}
	return SecretKeyFromBytes(bts)
}

// PublicKeyFromBytes creates a BLS public key from a  BigEndian byte slice.
func PublicKeyFromBytes(pubKey []byte) (PublicKey, error) {
	return currImplementation.PublicKeyFromBytes(pubKey)
}

// SignatureFromBytes creates a BLS signature from a LittleEndian byte slice.
func SignatureFromBytes(sig []byte) (Signature, error) {
	return currImplementation.SignatureFromBytes(sig)
}

// AggregatePublicKeys aggregates the provided raw public keys into a single key.
func AggregatePublicKeys(pubs [][]byte) (PublicKey, error) {
	return currImplementation.AggregatePublicKeys(pubs)
}

// AggregateSignatures converts a list of signatures into a single, aggregated sig.
func AggregateSignatures(sigs []common.Signature) common.Signature {
	return currImplementation.AggregateSignatures(sigs)
}

// VerifyMultipleSignatures verifies multiple signatures for distinct messages securely.
func VerifyMultipleSignatures(sigs [][]byte, msgs [][32]byte, pubKeys []common.PublicKey) (bool, error) {
	return currImplementation.VerifyMultipleSignatures(sigs, msgs, pubKeys)
}

// NewAggregateSignature creates a blank aggregate signature.
func NewAggregateSignature() common.Signature {
	return currImplementation.NewAggregateSignature()
}

// RandKey creates a new private key using a random input.
func RandKey() (common.SecretKey, error) {
	return currImplementation.RandKey()
}

// VerifyCompressed signature.
func VerifyCompressed(signature, pub, msg []byte) bool {
	return currImplementation.VerifyCompressed(signature, pub, msg)
}
