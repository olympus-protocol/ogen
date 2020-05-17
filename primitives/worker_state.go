package primitives

import (
	"io"

	"github.com/olympus-protocol/ogen/utils/serializer"
)

// WorkerStatus represents the status of a worker.
type WorkerStatus uint8

func (w WorkerStatus) String() string {
	switch w {
	case StatusActive:
		return "active"
	case StatusActivePendingExit:
		return "pending exit"
	case StatusExitedWithPenalty:
		return "penalty exit"
	case StatusExitedWithoutPenalty:
		return "exited"
	case StatusStarting:
		return "starting"
	default:
		return "unknown"
	}
}

const (
	// StatusStarting is when the validator is waiting to join.
	StatusStarting WorkerStatus = iota

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

// Worker is a worker in the queue.
type Worker struct {
	Balance          uint64
	PubKey           [48]byte
	PayeeAddress     [20]byte
	Status           WorkerStatus
	FirstActiveEpoch int64
	LastActiveEpoch  int64
}

// IsActive checks if a validator is currently active.
func (wr *Worker) IsActive() bool {
	return wr.Status == StatusActive || wr.Status == StatusActivePendingExit
}

// IsActiveAtEpoch checks if a validator is active at a slot.
func (wr *Worker) IsActiveAtEpoch(epoch int64) bool {
	return wr.IsActive() &&
		(wr.FirstActiveEpoch == -1 || wr.FirstActiveEpoch <= epoch) &&
		(wr.LastActiveEpoch == -1 || epoch <= wr.LastActiveEpoch)
}

// Serialize serializes a WorkerRow to the provided writer.
func (wr *Worker) Serialize(w io.Writer) error {
	return serializer.WriteElements(w, wr.Balance, wr.PubKey, wr.PayeeAddress, wr.Status, wr.FirstActiveEpoch, wr.LastActiveEpoch)
}

// Deserialize deserializes a worker row from the provided reader.
func (wr *Worker) Deserialize(r io.Reader) error {
	return serializer.ReadElements(r, &wr.Balance, &wr.PubKey, &wr.PayeeAddress, &wr.Status, &wr.FirstActiveEpoch, &wr.LastActiveEpoch)
}

// Copy returns a copy of the worker.
func (wr *Worker) Copy() Worker {
	return *wr
}
