package primitives

import (
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

type Worker struct {
	OutPoint     OutPoint
	Balance      uint64
	PubKey       [48]byte
	PayeeAddress string
}

// Serialize serializes a WorkerRow to the provided writer.
func (wr *Worker) Serialize(w io.Writer) error {
	err := wr.OutPoint.Serialize(w)
	if err != nil {
		return err
	}

	return serializer.WriteElements(w, wr.Balance, wr.PubKey, wr.PayeeAddress)
}

// Deserialize deserializes a worker row from the provided reader.
func (wr *Worker) Deserialize(r io.Reader) error {
	err := wr.OutPoint.Deserialize(r)
	if err != nil {
		return err
	}
	return serializer.ReadElements(r, &wr.Balance, &wr.PubKey, &wr.PayeeAddress)
}

type WorkerState struct {
	Workers map[chainhash.Hash]Worker
}

// Have checks if a Worker exists.
func (w *WorkerState) Have(c chainhash.Hash) bool {
	_, found := w.Workers[c]
	return found
}

// Get gets a Worker from state.
func (w *WorkerState) Get(c chainhash.Hash) Worker {
	return w.Workers[c]
}

func (w *WorkerState) Serialize(wr io.Writer) error {
	if err := serializer.WriteVarInt(wr, uint64(len(w.Workers))); err != nil {
		return err
	}

	for h, utxo := range w.Workers {
		if _, err := wr.Write(h[:]); err != nil {
			return err
		}

		if err := utxo.Serialize(wr); err != nil {
			return err
		}
	}

	return nil
}

func (w *WorkerState) Deserialize(r io.Reader) error {
	if w.Workers == nil {
		w.Workers = make(map[chainhash.Hash]Worker)
	}

	numWorkers, err := serializer.ReadVarInt(r)

	if err != nil {
		return err
	}

	for i := uint64(0); i < numWorkers; i++ {
		var hash chainhash.Hash
		if _, err := r.Read(hash[:]); err != nil {
			return err
		}

		var worker Worker
		if err := worker.Deserialize(r); err != nil {
			return err
		}

		w.Workers[hash] = worker
	}

	return nil
}
