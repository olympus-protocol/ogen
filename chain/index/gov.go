package index

import (
	"bytes"
	"io"
	"sync"

	"github.com/olympus-protocol/ogen/gov"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// GovRow is an item in the governance index.
type GovRow struct {
	OutPoint p2p.OutPoint
	GovData  gov.GovObject
}

// Serialize serializes a GovRow to the passed writer.
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

// Deserialize deserialized a GovRow from the passed reader.
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

// GovIndex maps gov row hashes to rows.
type GovIndex struct {
	lock  sync.RWMutex
	index map[chainhash.Hash]*GovRow
}

// Serialize serializes the GovIndex to the passed writer.
func (i *GovIndex) Serialize(w io.Writer) error {
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

// Deserialize deserializes a GovIndex from the passed reader.
func (i *GovIndex) Deserialize(r io.Reader) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	buf, _ := r.(*bytes.Buffer)
	if buf.Len() > 0 {
		count, err := serializer.ReadVarInt(r)
		if err != nil {
			return err
		}
		i.index = make(map[chainhash.Hash]*GovRow, count)
		for k := uint64(0); k < count; k++ {
			var row *GovRow
			err = row.Deserialize(r)
			if err != nil {
				return err
			}
			i.add(row)
		}
		return nil
	}
	return nil
}

// Get gets a row from the GovIndex
func (i *GovIndex) Get(hash chainhash.Hash) *GovRow {
	i.lock.RLock()
	defer i.lock.RUnlock()
	row, _ := i.index[hash]
	return row
}

// Have checks if the GovIndex contains a hash.
func (i *GovIndex) Have(hash chainhash.Hash) bool {
	i.lock.RLock()
	defer i.lock.RUnlock()
	_, ok := i.index[hash]
	return ok
}

func (i *GovIndex) add(row *GovRow) {
	i.index[row.GovData.GovID] = row
	return
}

// Add adds a row to the GovIndex.
func (i *GovIndex) Add(row *GovRow) {
	i.lock.Lock()
	i.add(row)
	i.lock.Unlock()
	return
}

// InitGovIndex initializes a gov index.
func InitGovIndex() *GovIndex {
	return &GovIndex{
		index: make(map[chainhash.Hash]*GovRow),
	}
}
