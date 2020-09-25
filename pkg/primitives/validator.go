package primitives

const (
	// StatusStarting is when the validator is waiting to join.
	StatusStarting uint64 = iota

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
	PubKey           [48]byte
	PayeeAddress     [20]byte
	Status           uint64
	FirstActiveEpoch uint64
	LastActiveEpoch  uint64
}

// StatusString returns the status on human readable string
func (v *Validator) StatusString() string {
	switch v.Status {
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

// IsActive checks if a validator is currently active.
func (v *Validator) IsActive() bool {
	return v.Status == StatusActive || v.Status == StatusActivePendingExit
}

// IsActiveAtEpoch checks if a validator is active at a slot.
func (v *Validator) IsActiveAtEpoch(epoch uint64) bool {
	return v.IsActive() &&
		(v.FirstActiveEpoch == 0 || v.FirstActiveEpoch <= epoch) &&
		(v.LastActiveEpoch == 0 || epoch <= v.LastActiveEpoch)
}

// Marshal encodes the data.
func (v *Validator) Marshal() ([]byte, error) {
	return v.MarshalSSZ()
}

// Unmarshal decodes the data.
func (v *Validator) Unmarshal(b []byte) error {
	return v.UnmarshalSSZ(b)
}

// Copy returns a copy of the validator.
func (v *Validator) Copy() Validator {
	return *v
}
