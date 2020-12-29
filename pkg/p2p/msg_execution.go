package p2p

// MsgExecution is the message that contains the locator to fetch blocks.
type MsgExecution struct {
	FromPubKey [48]byte
	To         [20]byte
	Input      []byte `ssz-max:"32768"`
	Signature  [96]byte
}

// Marshal serializes the data to bytes
func (m *MsgExecution) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgExecution) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgExecution) Command() string {
	return MsgExecutionCmd
}

// MaxPayloadLength returns the maximum size of the MsgGetBlocks message.
func (m *MsgExecution) MaxPayloadLength() uint64 {
	return 32940
}
