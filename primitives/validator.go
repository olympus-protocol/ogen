package primitives

import (
	"github.com/prysmaticlabs/go-ssz"
)

// ValidatorStatus represents the status of a Validator.
type ValidatorStatus uint8

func (w ValidatorStatus) String() string {
	switch w {
	case StatusActive:
		return "ACTIVE"
	case StatusActivePendingExit:
		return "PENDING_EXIT"
	case StatusExitedWithPenalty:
		return "PENALTY_EXIT"
	case StatusExitedWithoutPenalty:
		return "EXITED"
	case StatusStarting:
		return "STARTING"
	default:
		return "UNKNOWN"
	}
}

const (
	// StatusStarting is when the validator is waiting to join.
	StatusStarting ValidatorStatus = iota

	// StatusActive is when the validator is currently in the queue.
	StatusActive

	// StatusActivePendingExit is when a validator is queued to be removed from the
	// active set.
	StatusActivePendingExit

	// StatusExitedWithPenalty is when the validator is exited due to a slashing condition
	// being violated.
	StatusExitedWithPenalty

	// StatusExitedWithoutPenalty is when a validator is exited due to a drop below
	// the ejection balance.
	StatusExitedWithoutPenalty
)

// Validator is a validator in the queue.
type Validator struct {
	Balance          uint64
	PubKey           []byte
	PayeeAddress     [20]byte
	Status           ValidatorStatus
	FirstActiveEpoch int64
	LastActiveEpoch  int64
}

// IsActive checks if a validator is currently active.
func (wr *Validator) IsActive() bool {
	return wr.Status == StatusActive || wr.Status == StatusActivePendingExit
}

// IsActiveAtEpoch checks if a validator is active at a slot.
func (wr *Validator) IsActiveAtEpoch(epoch int64) bool {
	return wr.IsActive() &&
		(wr.FirstActiveEpoch == -1 || wr.FirstActiveEpoch <= epoch) &&
		(wr.LastActiveEpoch == -1 || epoch <= wr.LastActiveEpoch)
}

// Marshal encodes the data.
func (v *Validator) Marshal() ([]byte, error) {
	return ssz.Marshal(v)
}

// Unmarshal decodes the data.
func (v *Validator) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, v)
}

// Copy returns a copy of the validator.
func (wr *Validator) Copy() Validator {
	return *wr
}
