package p2p

// MsgGetAddr is the struct containing the getaddr message command.
type MsgGetAddr struct{}

// Marshal serializes the data to bytes
func (m *MsgGetAddr) Marshal() ([]byte, error) {
	return []byte{}, nil
}

// Unmarshal deserializes the data
func (m *MsgGetAddr) Unmarshal([]byte) error {
	return nil
}

// Command returns the message topic
func (m *MsgGetAddr) Command() string {
	return MsgGetAddrCmd
}

// MaxPayloadLength returns the maximum size of the MsgGetAddr message.
func (m *MsgGetAddr) MaxPayloadLength() uint64 {
	return 0
}
