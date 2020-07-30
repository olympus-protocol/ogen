// Package bls implements a go-wrapper around a library implementing the
// the bls-381 curve and signature scheme. This package exposes a public API for
// verifying and aggregating BLS signatures used by Ethereum 2.0.
package bls

import (
	"io"
	"math/big"

	"github.com/dgraph-io/ristretto"
	"github.com/olympus-protocol/bls-go/bls"
)

// KeyPair is an interface struct to serve keypairs
type KeyPair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

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

var sigCache, _ = ristretto.NewCache(&ristretto.Config{
	NumCounters: maxKeys,
	MaxCost:     1 << 19, // 500 kb is cache max size
	BufferItems: 64,
})

// RFieldModulus for the bls-381 curve.
var RFieldModulus, _ = new(big.Int).SetString("52435875175126190479447740508185965837690552500527637822603658699938581184513", 10)

// FunctionalPublicKey is either a multipub or a regular public key.
type FunctionalPublicKey interface {
	Marshal() []byte
	Unmarshal(b []byte) error
	Hash() ([20]byte, error)
	Type() FunctionalSignatureType
}

// FunctionalSignatureType is a functional signature type.
type FunctionalSignatureType uint8

const (
	// TypeSingle is a single signature.
	TypeSingle FunctionalSignatureType = iota

	// TypeMulti is a multisignature.
	TypeMulti
)

// FunctionalSignature is a signature that can be included in transactions or votes.
type FunctionalSignature interface {
	Marshal() ([]byte, error)
	Unmarshal(b []byte) error
	Sign(secKey *SecretKey, msg []byte) error
	Verify(msg []byte) bool
	GetPublicKey() (FunctionalPublicKey, error)
	Type() FunctionalSignatureType
	Copy() FunctionalSignature
}

// WriteFunctionalSignature writes any type of functional signature.
func WriteFunctionalSignature(w io.Writer, sig FunctionalSignature) error {
	if _, err := w.Write([]byte{byte(sig.Type())}); err != nil {
		return err
	}
	sigBytes, err := sig.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(sigBytes)
	if err != nil {
		return err
	}
	return nil
}

// ReadFunctionalSignature reads any type of functional signature.
func ReadFunctionalSignature(r io.Reader) (FunctionalSignature, error) {
	sigType := make([]byte, 1)
	if _, err := io.ReadFull(r, sigType); err != nil {
		return nil, err
	}

	var out FunctionalSignature
	switch FunctionalSignatureType(sigType[0]) {
	case TypeSingle:
		out = new(CombinedSignature)
	case TypeMulti:
		out = new(Multisig)
	}
	data := []byte{}
	_, err := io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}
	if err := out.Unmarshal(data); err != nil {
		return nil, err
	}

	return out, nil
}
