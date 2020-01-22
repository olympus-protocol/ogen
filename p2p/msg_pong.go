package p2p

import (
	"bytes"
	"fmt"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

type MsgPong struct {
	Nonce uint64
}

func (m *MsgPong) Encode(w io.Writer) error {
	err := serializer.WriteElements(w, m.Nonce)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgPong) Decode(r io.Reader) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("MsgPong.Decode reader is not a *bytes.Buffer")
	}
	if buf.Len() > 0 {
		err := serializer.ReadElement(buf, &m.Nonce)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MsgPong) Command() string {
	return MsgPongCmd
}

func (m *MsgPong) MaxPayloadLength() uint32 {
	return 8
}

func NewMsgPong(nonce uint64) *MsgPong {
	return &MsgPong{
		Nonce: nonce,
	}
}
