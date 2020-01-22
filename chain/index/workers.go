package index

import (
	"bytes"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"github.com/olympus-protocol/ogen/workers"
	"io"
	"sync"
)

type WorkerRow struct {
	OutPoint   p2p.OutPoint
	WorkerData workers.Worker
}

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

type WorkerIndex struct {
	Index map[chainhash.Hash]*WorkerRow
	lock  sync.RWMutex
}

func (i *WorkerIndex) Serialize(w io.Writer) error {
	err := serializer.WriteVarInt(w, uint64(len(i.Index)))
	if err != nil {
		return err
	}
	for _, row := range i.Index {
		err = row.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *WorkerIndex) Deserialize(r io.Reader) error {
	buf, _ := r.(*bytes.Buffer)
	if buf.Len() > 0 {
		count, err := serializer.ReadVarInt(r)
		if err != nil {
			return err
		}
		i.Index = make(map[chainhash.Hash]*WorkerRow, count)
		for k := uint64(0); k < count; k++ {
			var row *WorkerRow
			err = row.Deserialize(r)
			if err != nil {
				return err
			}
			err := i.Add(row)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func (i *WorkerIndex) Get(hash chainhash.Hash) *WorkerRow {
	i.lock.Lock()
	row, _ := i.Index[hash]
	i.lock.Unlock()
	return row
}

func (i *WorkerIndex) Have(hash chainhash.Hash) bool {
	i.lock.Lock()
	_, ok := i.Index[hash]
	i.lock.Unlock()
	return ok
}

func (i *WorkerIndex) Add(row *WorkerRow) error {
	buf := bytes.NewBuffer([]byte{})
	err := row.OutPoint.Serialize(buf)
	if err != nil {
		return err
	}
	workerIDHash := chainhash.DoubleHashH(buf.Bytes())
	i.lock.Lock()
	i.Index[workerIDHash] = row
	i.lock.Unlock()
	return nil
}

func InitWorkersIndex() *WorkerIndex {
	return &WorkerIndex{
		Index: make(map[chainhash.Hash]*WorkerRow),
	}
}
