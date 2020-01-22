package p2p

import "io"

type MsgGetAddr struct{}

func (m *MsgGetAddr) Encode(w io.Writer) error {
	return nil
}

func (m *MsgGetAddr) Decode(r io.Reader) error {
	return nil
}

func (m *MsgGetAddr) Command() string {
	return MsgGetAddrCmd
}

func (m *MsgGetAddr) MaxPayloadLength() uint32 {
	return 0
}

func NewMsgGetAddr() *MsgGetAddr {
	return &MsgGetAddr{}
}
