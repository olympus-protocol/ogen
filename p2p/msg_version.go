package p2p

import (
	"github.com/golang/snappy"
)

// MsgVersion is the struct that contains the node information during the version handshake.
type MsgVersion struct {
	LastBlock uint64 // 8 bytes
	Nonce     uint64 // 8 bytes
	Timestamp uint64 // 8 bytes
}

// Marshal serializes the data to bytes
func (m *MsgVersion) Marshal() ([]byte, error) {
	b, err := m.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if uint64(len(b)) > m.MaxPayloadLength() {
		return nil, ErrorSizeExceed
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal deserializes the data
func (m *MsgVersion) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if uint64(len(d)) > m.MaxPayloadLength() {
		return ErrorSizeExceed
	}
	return m.UnmarshalSSZ(d)
}

// Command returns the message topic
func (m *MsgVersion) Command() string {
	return MsgVersionCmd
}

// MaxPayloadLength returns the maximum size of the MsgVersion message.
func (m *MsgVersion) MaxPayloadLength() uint64 {
	return 24
}
