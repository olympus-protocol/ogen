package p2p

// MaxAddrPerMsg defines the maximum address that can be added into an addr message.
const MaxAddrPerMsg = 32

// MsgAddr is the struct for the response of getaddr.
type MsgAddr struct {
	Addr [][256]byte `ssz-max:"32"`
}

// Marshal serializes the data to bytes
func (m *MsgAddr) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgAddr) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgAddr) Command() string {
	return MsgAddrCmd
}

// MaxPayloadLength returns the maximum size of the MsgAddr message.
func (m *MsgAddr) MaxPayloadLength() uint64 {
	return uint64(MaxAddrPerMsg*512) + 4
}
