package index

import (
	"bytes"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/users"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
	"sync"
)

type UserRow struct {
	OutPoint p2p.OutPoint
	UserData users.User
}

func (ur *UserRow) Serialize(w io.Writer) error {
	err := ur.OutPoint.Serialize(w)
	if err != nil {
		return err
	}
	err = ur.UserData.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

func (ur *UserRow) Deserialize(r io.Reader) error {
	err := ur.OutPoint.Deserialize(r)
	if err != nil {
		return err
	}
	err = ur.UserData.Deserialize(r)
	if err != nil {
		return err
	}
	return nil
}

type UserIndex struct {
	lock  sync.RWMutex
	Index map[chainhash.Hash]*UserRow
}

func (i *UserIndex) Serialize(w io.Writer) error {
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

func (i *UserIndex) Deserialize(r io.Reader) error {
	buf, _ := r.(*bytes.Buffer)
	if buf.Len() > 0 {
		count, err := serializer.ReadVarInt(r)
		if err != nil {
			return err
		}
		i.Index = make(map[chainhash.Hash]*UserRow, count)
		for k := uint64(0); k < count; k++ {
			var row *UserRow
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

func (i *UserIndex) Get(hash chainhash.Hash) *UserRow {
	i.lock.Lock()
	row, _ := i.Index[hash]
	i.lock.Unlock()
	return row
}

func (i *UserIndex) Have(hash chainhash.Hash) bool {
	i.lock.Lock()
	_, ok := i.Index[hash]
	i.lock.Unlock()
	return ok
}

func (i *UserIndex) Add(row *UserRow) {
	userHash := chainhash.DoubleHashH([]byte(row.UserData.Name))
	i.lock.Lock()
	i.Index[userHash] = row
	i.lock.Unlock()
	return
}

func InitUsersIndex() *UserIndex {
	return &UserIndex{
		Index: make(map[chainhash.Hash]*UserRow),
	}
}
