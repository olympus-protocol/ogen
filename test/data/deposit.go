package testdata

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/primitives"
)

var randPrv = bls.RandKey()

var DepositData = primitives.DepositData{
	PublicKey:         randPrv.PublicKey().Marshal(),
	ProofOfPossession: sig.Marshal(),
	WithdrawalAddress: [20]byte{0x0, 0x1, 0x2},
}

var Deposit = primitives.Deposit{
	PublicKey: randPrv.PublicKey().Marshal(),
	Signature: sig.Marshal(),
	Data:      DepositData,
}
