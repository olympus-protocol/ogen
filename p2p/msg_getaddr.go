package p2p

type MsgGetAddr struct{}

// Marshal serializes the data to bytes
func (m *MsgGetAddr) Marshal() ([]byte, error) {
	return []byte{}, nil
}

// Unmarshal deserializes the data
func (m *MsgGetAddr) Unmarshal([]byte) error {
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
