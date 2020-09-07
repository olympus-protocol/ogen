package p2p

// MsgVersion is the struct that contains the node information during the version handshake.
type MsgVersion struct {
	TipSlot            uint64   // 8 bytes
	Nonce              uint64   // 8 bytes
	Timestamp          uint64   // 8 bytes
	LastJustifiedHash  [32]byte // 32 bytes
	LastJustifiedEpoch uint64   // 8 bytes
}

// Marshal serializes the data to bytes
func (m *MsgVersion) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgVersion) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgVersion) Command() string {
	return MsgVersionCmd
}

// MaxPayloadLength returns the maximum size of the MsgVersion message.
func (m *MsgVersion) MaxPayloadLength() uint64 {
	return 64
}
