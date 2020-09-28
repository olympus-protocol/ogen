package p2p

// MsgFinalization is the struct that contains the node information when announcing a finalization.
type MsgFinalization struct {
	Tip             uint64
	TipSlot         uint64
	TipHash         [32]byte
	JustifiedSlot   uint64
	JustifiedHeight uint64
	JustifiedHash   [32]byte
	FinalizedSlot   uint64
	FinalizedHeight uint64
	FinalizedHash   [32]byte
}

// Marshal serializes the data to bytes
func (m *MsgFinalization) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgFinalization) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgFinalization) Command() string {
	return MsgFinalizationCmd
}

// MaxPayloadLength returns the maximum size of the MsgVersion message.
func (m *MsgFinalization) MaxPayloadLength() uint64 {
	return 144
}
