package primitives

const (
	// StatusStarting is when the validator is waiting to join.
	StatusStarting uint8 = iota

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
	PubKey           []byte `ssz-size:"48"`
	PayeeAddress     []byte `ssz-size:"20"`
	Status           uint8
	FirstActiveEpoch uint64
	LastActiveEpoch  uint64
}

// IsActive checks if a validator is currently active.
func (wr *Validator) IsActive() bool {
	return wr.Status == StatusActive || wr.Status == StatusActivePendingExit
}

// IsActiveAtEpoch checks if a validator is active at a slot.
func (wr *Validator) IsActiveAtEpoch(epoch uint64) bool {
	return wr.IsActive() &&
		(wr.FirstActiveEpoch == 0 || wr.FirstActiveEpoch <= epoch) &&
		(wr.LastActiveEpoch == 0 || epoch <= wr.LastActiveEpoch)
}

// Copy returns a copy of the validator.
func (wr *Validator) Copy() Validator {
	return *wr
}
