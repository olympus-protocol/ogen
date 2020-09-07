package p2p

// MsgSyncEnd is the struct containing the getaddr message command.
type MsgSyncEnd struct{}

// Marshal serializes the data to bytes
func (m *MsgSyncEnd) Marshal() ([]byte, error) {
	return []byte{}, nil
}

// Unmarshal deserializes the data
func (m *MsgSyncEnd) Unmarshal([]byte) error {
	return nil
}

// Command returns the message topic
func (m *MsgSyncEnd) Command() string {
	return MsgGetAddrCmd
}

// MaxPayloadLength returns the maximum size of the MsgSyncEnd message.
func (m *MsgSyncEnd) MaxPayloadLength() uint64 {
	return 0
}
