package users

import (
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

type User struct {
	PubKey [48]byte
	Name   string
}

func (u *User) Serialize(w io.Writer) error {
	err := serializer.WriteElement(w, u.PubKey)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, u.Name)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) Deserialize(r io.Reader) error {
	err := serializer.ReadElement(r, &u.PubKey)
	if err != nil {
		return err
	}
	u.Name, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	return nil
}
