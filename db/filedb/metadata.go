package filedb

import (
	"github.com/grupokindynos/ogen/utils/serializer"
	"io"
)

type MetaData struct {
	Version     int64
	Timestamp   int64
	MaxElemSize int64
	Name        string
}

func (m *MetaData) serialize(w io.Writer) (err error) {
	err = serializer.WriteVarInt(w, m.serializedSize())
	if err != nil {
		return err
	}
	err = serializer.WriteElements(w, m.Version, m.Timestamp, m.MaxElemSize)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, m.Name)
	if err != nil {
		return err
	}
	return nil
}

func (m *MetaData) deserialize(r io.Reader) (err error) {
	_, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	err = serializer.ReadElements(r, &m.Version, &m.Timestamp, &m.MaxElemSize)
	if err != nil {
		return err
	}
	m.Name, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	return nil
}

func (m *MetaData) serializedSize() uint64 {
	/*
		[8] Size
		[8] Version
		[8] Timestamp
		[8] MaxElemSize
		[8] string size + [len(m.Name) Name
	*/
	return uint64(40 + len(m.Name))
}
