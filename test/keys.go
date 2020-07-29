package testdata

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/bitfield"
)

var sigB = [96]byte{}
var pubB = [48]byte{}

var randKey = bls.RandKey().PublicKey()
var randKeyBytes = bls.RandKey().PublicKey().Marshal()

var sig = bls.NewAggregateSignature()
var funcSigByte, _ = CombinedSignature.Marshal()
var sigBytes = bls.NewAggregateSignature().Marshal()

func init() {
	copy(sigB[:], sigBytes)
	copy(pubB[:], randKeyBytes)
}

var CombinedSignature = bls.CombinedSignature{
	S: sigB,
	P: pubB,
}

var Multipub = bls.Multipub{
	PublicKeys: [][48]byte{pubB, pubB, pubB, pubB, pubB},
	NumNeeded:  5,
}

var Multisig = bls.Multisig{
	PublicKey:  &Multipub,
	Signatures: [][96]byte{sigB, sigB, sigB, sigB, sigB, sigB},
	KeysSigned: bitfield.NewBitfield(6),
}
