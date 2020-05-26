// Package bls implements a go-wrapper around a library implementing the
// the bls-381 curve and signature scheme. This package exposes a public API for
// verifying and aggregating BLS signatures used by Ethereum 2.0.
package bls

import (
	"math/big"

	"github.com/dgraph-io/ristretto"
	"github.com/olympus-protocol/bls-go/bls"
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
