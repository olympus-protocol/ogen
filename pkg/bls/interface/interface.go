package bls_interface

import (
	"github.com/olympus-protocol/ogen/pkg/params"
	"math/big"
)

// Prefix is a global variable for prefixes used for account bech32 encoding.
// Set to mainnet if not override by the Initialize function.
var Prefix params.AccountPrefixes = params.Mainnet.AccountPrefixes

// KeyPair is an interface struct to serve keypairs
type KeyPair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

// SecretKey represents a BLS secret or private key.
type SecretKey interface {
	PublicKey() PublicKey
	Sign(msg []byte) Signature
	Marshal() []byte
	ToWIF() string
}

// PublicKey represents a BLS public key.
type PublicKey interface {
	Marshal() []byte
	Copy() PublicKey
	Aggregate(p2 PublicKey) PublicKey
	Hash() ([20]byte, error)
	ToAccount() string
}

// Signature represents a BLS signature.
type Signature interface {
	Verify(pubKey PublicKey, msg []byte) bool
	AggregateVerify(pubKeys []PublicKey, msgs [][32]byte) bool
	FastAggregateVerify(pubKeys []PublicKey, msg [32]byte) bool
	Marshal() []byte
	Copy() Signature
}

// RFieldModulus for the bls-381 curve.
var RFieldModulus, _ = new(big.Int).SetString("52435875175126190479447740508185965837690552500527637822603658699938581184513", 10)
