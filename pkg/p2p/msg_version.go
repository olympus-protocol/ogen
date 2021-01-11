package p2p

// MsgVersion is the struct that contains the node information during the version handshake.
type MsgVersion struct {
	Tip             uint64
	TipSlot         uint64
	TipHash         [32]byte
	Nonce           uint64
	Timestamp       uint64
	JustifiedSlot   uint64
	JustifiedHeight uint64
	JustifiedHash   [32]byte
	FinalizedSlot   uint64
	FinalizedHeight uint64
	FinalizedHash   [32]byte
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
	return 240
}

// PayloadLength returns the size of the MsgVersion message.
func (m *MsgVersion) PayloadLength() uint64 {
	return uint64(m.SizeSSZ())
}
