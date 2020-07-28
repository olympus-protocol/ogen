package testdata

import (
	"github.com/olympus-protocol/ogen/primitives"
)

var DepositData = primitives.DepositData{
	PublicKey:         pubB,
	ProofOfPossession: sigB,
	WithdrawalAddress: [20]byte{0x0, 0x1, 0x2},
}

var Deposit = primitives.Deposit{
	PublicKey: pubB,
	Signature: sigB,
	Data:      &DepositData,
}
