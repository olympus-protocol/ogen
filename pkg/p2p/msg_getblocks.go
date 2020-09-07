package p2p

// MsgGetBlocks is the message that contains the locator to fetch blocks.
type MsgGetBlocks struct {
	LastBlockHash [32]byte
}

// Marshal serializes the data to bytes
func (m *MsgGetBlocks) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgGetBlocks) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgGetBlocks) Command() string {
	return MsgGetBlocksCmd
}

// MaxPayloadLength returns the maximum size of the MsgGetBlocks message.
func (m *MsgGetBlocks) MaxPayloadLength() uint64 {
	return 32
}
