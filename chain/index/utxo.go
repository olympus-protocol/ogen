package index

import (
	"bytes"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
	"sync"
)

type UtxoRow struct {
	OutPoint          p2p.OutPoint
	PrevInputsPubKeys [][48]byte
	Owner             string
	Amount            int64
}

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

type UtxosIndex struct {
	lock  sync.Mutex
	Index map[chainhash.Hash]*UtxoRow
}

func (i *UtxosIndex) Serialize(w io.Writer) error {
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

func (i *UtxosIndex) Deserialize(r io.Reader) error {
	buf, _ := r.(*bytes.Buffer)
	if buf.Len() > 0 {
		count, err := serializer.ReadVarInt(r)
		if err != nil {
			return err
		}
		i.Index = make(map[chainhash.Hash]*UtxoRow, count)
		for k := uint64(0); k < count; k++ {
			var row *UtxoRow
			err = row.Deserialize(r)
			if err != nil {
				return err
			}
			err = i.Add(row)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func (i *UtxosIndex) Get(hash chainhash.Hash) *UtxoRow {
	i.lock.Lock()
	row, _ := i.Index[hash]
	i.lock.Unlock()
	return row
}

func (i *UtxosIndex) Have(hash chainhash.Hash) bool {
	i.lock.Lock()
	_, ok := i.Index[hash]
	i.lock.Unlock()
	return ok
}

func (i *UtxosIndex) Add(row *UtxoRow) error {
	buf := bytes.NewBuffer([]byte{})
	err := row.OutPoint.Serialize(buf)
	if err != nil {
		return err
	}
	utxoHash := chainhash.DoubleHashH(buf.Bytes())
	i.lock.Lock()
	i.Index[utxoHash] = row
	i.lock.Unlock()
	return nil
}

func InitUtxosIndex() *UtxosIndex {
	return &UtxosIndex{
		Index: make(map[chainhash.Hash]*UtxoRow),
	}
}
