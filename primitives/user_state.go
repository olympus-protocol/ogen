package primitives

import (
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

type User struct {
	OutPoint OutPoint
	PubKey   [48]byte
	Name     string
}

// Serialize serializes the UserRow to a writer.
func (ur *User) Serialize(w io.Writer) error {
	err := ur.OutPoint.Serialize(w)
	if err != nil {
		return err
	}
	if err := serializer.WriteElements(w, ur.PubKey, ur.Name); err != nil {
		return err
	}
	return nil
}

// Deserialize deserializes a user from the writer.
func (ur *User) Deserialize(r io.Reader) error {
	err := ur.OutPoint.Deserialize(r)
	if err != nil {
		return err
	}
	if err := serializer.ReadElements(r, &ur.PubKey, &ur.Name); err != nil {
		return err
	}
	return nil
}

// Copy returns a copy of the user.
func (ur *User) Copy() User {
	return *ur
}

// Hash gets the hash of the user.
func (ur *User) Hash() chainhash.Hash {
	return chainhash.DoubleHashH([]byte(ur.Name))
}

// UserState is the state of the user registry.
type UserState struct {
	Users map[chainhash.Hash]User
}

// Copy returns a copy of the user state.
func (u *UserState) Copy() UserState {
	newU := *u
	newU.Users = make(map[chainhash.Hash]User)

	for i, u := range newU.Users {
		newU.Users[i] = u.Copy()
	}

	return newU
}

func (u *UserState) Serialize(w io.Writer) error {
	if err := serializer.WriteVarInt(w, uint64(len(u.Users))); err != nil {
		return err
	}

	for h, user := range u.Users {
		if _, err := w.Write(h[:]); err != nil {
			return err
		}

		if err := user.Serialize(w); err != nil {
			return err
		}
	}

	return nil
}

func (u *UserState) Deserialize(r io.Reader) error {
	if u.Users == nil {
		u.Users = make(map[chainhash.Hash]User)
	}

	numUsers, err := serializer.ReadVarInt(r)

	if err != nil {
		return err
	}

	for i := uint64(0); i < numUsers; i++ {
		var hash chainhash.Hash
		if _, err := r.Read(hash[:]); err != nil {
			return err
		}

		var user User
		if err := user.Deserialize(r); err != nil {
			return err
		}

		u.Users[hash] = user
	}

	return nil
}

// Have checks if a User exists.
func (u *UserState) Have(c chainhash.Hash) bool {
	_, found := u.Users[c]
	return found
}

// Get gets a User from state.
func (u *UserState) Get(c chainhash.Hash) User {
	return u.Users[c]
}
