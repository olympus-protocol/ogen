package serializer

import (
	"bytes"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"io"
)

type Serializable interface {
	Encode(w io.Writer) error
	Decode(r io.Reader) error
}

func Hash(s Serializable) *chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})

	_ = s.Encode(buf)

	h := chainhash.HashH(buf.Bytes())
	return &h
}