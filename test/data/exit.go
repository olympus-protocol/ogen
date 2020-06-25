package testdata

import (
	"github.com/olympus-protocol/ogen/primitives"
)

var Exit = primitives.Exit{
	ValidatorPubkey: randPrv.PublicKey().Marshal(),
	WithdrawPubkey:  randPrv.PublicKey().Marshal(),
	Signature:       sig.Marshal(),
}
