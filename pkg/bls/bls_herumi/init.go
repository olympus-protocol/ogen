package bls_herumi

import (
	"github.com/dgraph-io/ristretto"
	bls12 "github.com/olympus-protocol/bls-go/bls"
)

type HerumiImplementation struct{}

func init() {
	if err := bls12.Init(bls12.BLS12_381); err != nil {
		panic(err)
	}
	if err := bls12.SetETHmode(bls12.EthModeDraft07); err != nil {
		panic(err)
	}
}

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
