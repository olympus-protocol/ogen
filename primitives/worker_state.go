package primitives

import (
	"io"

	"github.com/olympus-protocol/ogen/utils/serializer"
)

// WorkerStatus represents the status of a worker.
type WorkerStatus uint8

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
	Balance      uint64
	PubKey       [48]byte
	PayeeAddress [20]byte
	Status       WorkerStatus
}

// IsActive checks if a validator is currently active.
func (wr *Worker) IsActive() bool {
	return wr.Status == StatusActive || wr.Status == StatusActivePendingExit
}

// Serialize serializes a WorkerRow to the provided writer.
func (wr *Worker) Serialize(w io.Writer) error {
	return serializer.WriteElements(w, wr.Balance, wr.PubKey, wr.PayeeAddress, wr.Status)
}

// Deserialize deserializes a worker row from the provided reader.
func (wr *Worker) Deserialize(r io.Reader) error {
	return serializer.ReadElements(r, &wr.Balance, &wr.PubKey, &wr.PayeeAddress, &wr.Status)
}

// Copy returns a copy of the worker.
func (wr *Worker) Copy() Worker {
	return *wr
}
