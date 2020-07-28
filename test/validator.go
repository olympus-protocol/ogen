package testdata

import (
	"github.com/olympus-protocol/ogen/primitives"
)

var Validator = primitives.Validator{
	Balance:          100000,
	PubKey:           pubB,
	PayeeAddress:     [20]byte{0x0, 0x1, 0x9},
	Status:           primitives.StatusActive,
	FirstActiveEpoch: 0,
	LastActiveEpoch:  0,
}
