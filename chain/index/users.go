package index

import (
	"bytes"
	"io"
	"sync"

	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/users"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// UserRow represents a user in the UserIndex.
type UserRow struct {
	OutPoint p2p.OutPoint
	UserData users.User
}

// Serialize serializes the UserRow to a writer.
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

// Deserialize deserializes a user from the writer.
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

// UserIndex represents a map from hashes to users.
type UserIndex struct {
	lock  sync.RWMutex
	index map[chainhash.Hash]*UserRow
}

// Serialize serializes the user to the specified writer.
func (i *UserIndex) Serialize(w io.Writer) error {
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

// Deserialize deserializes the user index from the specified reader.
func (i *UserIndex) Deserialize(r io.Reader) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	buf, _ := r.(*bytes.Buffer)
	if buf.Len() > 0 {
		count, err := serializer.ReadVarInt(r)
		if err != nil {
			return err
		}
		i.index = make(map[chainhash.Hash]*UserRow, count)
		for k := uint64(0); k < count; k++ {
			var row *UserRow
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

// Get gets a user row from the index.
func (i *UserIndex) Get(hash chainhash.Hash) *UserRow {
	i.lock.RLock()
	row, _ := i.index[hash]
	i.lock.RUnlock()
	return row
}

// Have checks if the user index contains a row.
func (i *UserIndex) Have(hash chainhash.Hash) bool {
	i.lock.RLock()
	_, ok := i.index[hash]
	i.lock.RUnlock()
	return ok
}

func (i *UserIndex) add(row *UserRow) {
	userHash := chainhash.DoubleHashH([]byte(row.UserData.Name))
	i.index[userHash] = row
}

// Add adds a user row to the index.
func (i *UserIndex) Add(row *UserRow) {
	i.lock.Lock()
	i.add(row)
	i.lock.Unlock()
	return
}

// InitUsersIndex initializes the user index.
func InitUsersIndex() *UserIndex {
	return &UserIndex{
		index: make(map[chainhash.Hash]*UserRow),
	}
}
