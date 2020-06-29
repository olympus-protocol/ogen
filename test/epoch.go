package testdata

import (
	"github.com/olympus-protocol/ogen/primitives"
)

var EpochReceipt = primitives.EpochReceipt{
	Type:      primitives.PenaltyInactivityLeak,
	Amount:    100,
	Validator: 100,
}
