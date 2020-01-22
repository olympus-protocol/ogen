package p2p

import (
	"bytes"
	"fmt"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

type MsgPing struct {
	Nonce uint64
}

func (m *MsgPing) Encode(w io.Writer) error {
	err := serializer.WriteElements(w, m.Nonce)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgPing) Decode(r io.Reader) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("MsgPing.Decode reader is not a " +
			"*bytes.Buffer")
	}
	if buf.Len() > 0 {
		err := serializer.ReadElements(buf, &m.Nonce)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MsgPing) Command() string {
	return MsgPingCmd
}

func (m *MsgPing) MaxPayloadLength() uint32 {
	return 8
}

func NewMsgPing() *MsgPing {
	nonce, _ := serializer.RandomUint64()
	return &MsgPing{
		Nonce: nonce,
	}
}
