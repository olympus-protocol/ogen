package state

// Validator is a validator in the queue.
type Validator struct {
	Balance          uint64
	PubKey           []byte `ssz-size:"48"`
	PayeeAddress     []byte `ssz-size:"20"`
	Status           uint8
	FirstActiveEpoch uint64
	LastActiveEpoch  uint64
}
