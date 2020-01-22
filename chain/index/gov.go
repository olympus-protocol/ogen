package index

import (
	"bytes"
	"github.com/olympus-protocol/ogen/gov"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
	"sync"
)

type GovRow struct {
	OutPoint p2p.OutPoint
	GovData  gov.GovObject
}

func (gr *GovRow) Serialize(w io.Writer) error {
	err := gr.OutPoint.Serialize(w)
	if err != nil {
		return err
	}
	err = gr.GovData.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

func (gr *GovRow) Deserialize(r io.Reader) error {
	err := gr.OutPoint.Deserialize(r)
	if err != nil {
		return err
	}
	err = gr.GovData.Deserialize(r)
	if err != nil {
		return err
	}
	return nil
}

type GovIndex struct {
	lock  sync.RWMutex
	Index map[chainhash.Hash]*GovRow
}

func (i *GovIndex) Serialize(w io.Writer) error {
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

func (i *GovIndex) Deserialize(r io.Reader) error {
	buf, _ := r.(*bytes.Buffer)
	if buf.Len() > 0 {
		count, err := serializer.ReadVarInt(r)
		if err != nil {
			return err
		}
		i.Index = make(map[chainhash.Hash]*GovRow, count)
		for k := uint64(0); k < count; k++ {
			var row *GovRow
			err = row.Deserialize(r)
			if err != nil {
				return err
			}
			i.Add(row)
		}
		return nil
	}
	return nil
}

func (i *GovIndex) Get(hash chainhash.Hash) *GovRow {
	i.lock.Lock()
	row, _ := i.Index[hash]
	i.lock.Unlock()
	return row
}

func (i *GovIndex) Have(hash chainhash.Hash) bool {
	i.lock.Lock()
	_, ok := i.Index[hash]
	i.lock.Unlock()
	return ok
}

func (i *GovIndex) Add(row *GovRow) {
	i.lock.Lock()
	i.Index[row.GovData.GovID] = row
	i.lock.Unlock()
	return
}

func InitGovIndex() *GovIndex {
	return &GovIndex{
		Index: make(map[chainhash.Hash]*GovRow),
	}
}
