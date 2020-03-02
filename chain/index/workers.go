package index

import (
	"bytes"
	"io"
	"sync"

	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"github.com/olympus-protocol/ogen/workers"
)

// WorkerRow represent a worker row in the WorkerIndex.
type WorkerRow struct {
	OutPoint   p2p.OutPoint
	WorkerData workers.Worker
}

// Serialize serializes a WorkerRow to the provided writer.
func (wr *WorkerRow) Serialize(w io.Writer) error {
	err := wr.OutPoint.Serialize(w)
	if err != nil {
		return err
	}
	err = wr.WorkerData.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

// Deserialize deserializes a worker row from the provided reader.
func (wr *WorkerRow) Deserialize(r io.Reader) error {
	err := wr.OutPoint.Deserialize(r)
	if err != nil {
		return err
	}
	err = wr.WorkerData.Deserialize(r)
	if err != nil {
		return err
	}
	return nil
}

// WorkerIndex represents a map from worker hash to row.
type WorkerIndex struct {
	index map[chainhash.Hash]*WorkerRow
	lock  sync.RWMutex
}

// Serialize serializes the worker index to a writer.
func (i *WorkerIndex) Serialize(w io.Writer) error {
	i.lock.RLock()
	defer i.lock.RUnlock()
	err := serializer.WriteVarInt(w, uint64(len(i.index)))
	if err != nil {
		return err
	}
	for _, row := range i.index {
		err = row.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Deserialize deserializes the worker index from the provided reader.
func (i *WorkerIndex) Deserialize(r io.Reader) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	buf, _ := r.(*bytes.Buffer)
	if buf.Len() > 0 {
		count, err := serializer.ReadVarInt(r)
		if err != nil {
			return err
		}
		i.index = make(map[chainhash.Hash]*WorkerRow, count)
		for k := uint64(0); k < count; k++ {
			var row *WorkerRow
			err = row.Deserialize(r)
			if err != nil {
				return err
			}
			err := i.add(row)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

// Get gets a row from the worker index.
func (i *WorkerIndex) Get(hash chainhash.Hash) *WorkerRow {
	i.lock.RLock()
	row, _ := i.index[hash]
	i.lock.RUnlock()
	return row
}

// Have checks if a row exists in the worker index.
func (i *WorkerIndex) Have(hash chainhash.Hash) bool {
	i.lock.RLock()
	_, ok := i.index[hash]
	i.lock.RUnlock()
	return ok
}

func (i *WorkerIndex) add(row *WorkerRow) error {
	buf := bytes.NewBuffer([]byte{})
	err := row.OutPoint.Serialize(buf)
	if err != nil {
		return err
	}
	workerIDHash := chainhash.DoubleHashH(buf.Bytes())
	i.index[workerIDHash] = row
	return nil
}

// Add adds a row to the worker index.
func (i *WorkerIndex) Add(row *WorkerRow) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.add(row)
}

// InitWorkersIndex creates a new worker index.
func InitWorkersIndex() *WorkerIndex {
	return &WorkerIndex{
		index: make(map[chainhash.Hash]*WorkerRow),
	}
}
