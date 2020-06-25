package testdata

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/bitfield"
)

var randKey = bls.RandKey().PublicKey()
var randKeyBytes = bls.RandKey().PublicKey().Marshal()

var sig = bls.NewAggregateSignature()
var funcSigByte, _ = CombinedSignature.Marshal()
var sigBytes = bls.NewAggregateSignature().Marshal()

var CombinedSignature = bls.CombinedSignature{
	S: sig.Marshal(),
	P: randKey.Marshal(),
}

var Multipub = bls.Multipub{
	PublicKeys: [][]byte{randKeyBytes, randKeyBytes, randKeyBytes, randKeyBytes, randKeyBytes},
	NumNeeded: 5,
}

var Multisig = bls.Multisig{
	PublicKey: Multipub,
	Signatures: [][]byte{sigBytes,sigBytes,sigBytes,sigBytes,sigBytes,sigBytes},
	KeysSigned: bitfield.NewBitfield(5),
}