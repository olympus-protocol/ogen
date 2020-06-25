package testdata

import "github.com/olympus-protocol/ogen/bls"

var randKey = bls.RandKey().PublicKey()

var sig = bls.NewAggregateSignature()

var funcSig = bls.CombinedSignature{
	S: sig.Marshal(),
	P: randKey.Marshal(),
}

var funcSigByte, _ = funcSig.Marshal()
