package p2p

import "io"

type MsgVerack struct{}

func (m *MsgVerack) Encode(w io.Writer) error {
	return nil
}

func (m *MsgVerack) Decode(r io.Reader) error {
	return nil
}

func (m *MsgVerack) Command() string {
	return MsgVerackCmd
}

func (m *MsgVerack) MaxPayloadLength() uint32 {
	return 0
}

func NewMsgVerack() *MsgVerack {
	return &MsgVerack{}
}
