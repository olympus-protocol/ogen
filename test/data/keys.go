package testdata

import "github.com/olympus-protocol/ogen/bls"

var randKey = bls.RandKey().PublicKey()

var sig = bls.NewAggregateSignature()
