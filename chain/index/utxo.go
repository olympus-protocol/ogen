package index

import (
	"bytes"
	"io"
	"sync"

	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// UtxoRow is a Utxo in the UtxoIndex.
type UtxoRow struct {
	OutPoint          p2p.OutPoint
	PrevInputsPubKeys [][48]byte
	Owner             string
	Amount            int64
}

// Serialize serializes the UtxoRow to a writer.
func (l *UtxoRow) Serialize(w io.Writer) error {
	err := l.OutPoint.Serialize(w)
	if err != nil {
		return err
	}
	err = serializer.WriteVarInt(w, uint64(len(l.PrevInputsPubKeys)))
	if err != nil {
		return err
	}
	err = serializer.WriteElements(w, l.PrevInputsPubKeys)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, l.Owner)
	if err != nil {
		return err
	}
	err = serializer.WriteElement(w, l.Amount)
	if err != nil {
		return err
	}
	return nil
}

// Deserialize deserializes a UtxoRow from a reader.
func (l *UtxoRow) Deserialize(r io.Reader) error {
	err := l.OutPoint.Deserialize(r)
	if err != nil {
		return err
	}
	l.Owner, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	count, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	l.PrevInputsPubKeys = make([][48]byte, count)
	for i := uint64(0); i < count; i++ {
		var pubKey [48]byte
		err = serializer.ReadElement(r, &pubKey)
		if err != nil {
			return err
		}
		l.PrevInputsPubKeys = append(l.PrevInputsPubKeys, pubKey)
	}
	err = serializer.ReadElement(r, &l.Amount)
	if err != nil {
		return err
	}
	return nil
}

// UtxosIndex is an index mapping Utxo hashes to UtxoRow's.
type UtxosIndex struct {
	lock  sync.RWMutex
	index map[chainhash.Hash]*UtxoRow
}

// Serialize serializes the UtxosIndex to the specified writer.
func (i *UtxosIndex) Serialize(w io.Writer) error {
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

// Deserialize deserializes a UtxosIndex from a reader.
func (i *UtxosIndex) Deserialize(r io.Reader) error {
	i.lock.RLock()
	defer i.lock.RUnlock()
	buf, _ := r.(*bytes.Buffer)
	if buf.Len() > 0 {
		count, err := serializer.ReadVarInt(r)
		if err != nil {
			return err
		}
		i.index = make(map[chainhash.Hash]*UtxoRow, count)
		for k := uint64(0); k < count; k++ {
			var row *UtxoRow
			err = row.Deserialize(r)
			if err != nil {
				return err
			}
			err = i.add(row)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

// Get gets a row from the UtxoIndex by hash.
func (i *UtxosIndex) Get(hash chainhash.Hash) *UtxoRow {
	i.lock.RLock()
	row, _ := i.index[hash]
	i.lock.RUnlock()
	return row
}

// Have checks if a hash exists in the UtxosIndex.
func (i *UtxosIndex) Have(hash chainhash.Hash) bool {
	i.lock.RLock()
	_, ok := i.index[hash]
	i.lock.RUnlock()
	return ok
}

func (i *UtxosIndex) add(row *UtxoRow) error {
	buf := bytes.NewBuffer([]byte{})
	err := row.OutPoint.Serialize(buf)
	if err != nil {
		return err
	}
	utxoHash := chainhash.DoubleHashH(buf.Bytes())
	i.index[utxoHash] = row
	return nil
}

// Add adds a row to the UtxoIndex.
func (i *UtxosIndex) Add(row *UtxoRow) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.add(row)
}

// InitUtxosIndex initializes a new UtxosIndex.
func InitUtxosIndex() *UtxosIndex {
	return &UtxosIndex{
		index: make(map[chainhash.Hash]*UtxoRow),
	}
}
