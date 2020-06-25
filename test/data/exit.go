package testdata

import (
	"github.com/olympus-protocol/ogen/primitives"
)

var Exit = primitives.Exit{
	ValidatorPubkey: randKey.Marshal(),
	WithdrawPubkey:  randKey.Marshal(),
	Signature:       sig.Marshal(),
}
