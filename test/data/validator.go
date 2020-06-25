package testdata

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/primitives"
)

var randKey = bls.RandKey().PublicKey()
var Validator = primitives.Validator{
	Balance:          100000,
	PubKey:           randKey.Marshal(),
	PayeeAddress:     [20]byte{0x0, 0x1, 0x9},
	Status:           primitives.StatusActive,
	FirstActiveEpoch: 0,
	LastActiveEpoch:  0,
}
