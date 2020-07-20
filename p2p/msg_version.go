package p2p

import (
	"github.com/golang/snappy"
	"github.com/prysmaticlabs/go-ssz"
)

// MsgVersion is the struct that contains the node information during the version handshake.
type MsgVersion struct {
	LastBlock uint64 // 8 bytes
	Nonce     uint64 // 8 bytes
	Timestamp uint64 // 8 bytes
}

// Marshal serializes the data to bytes
func (m *MsgVersion) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(m)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal deserializes the data
func (m *MsgVersion) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(d, m)
}

// Command returns the message topic
func (m *MsgVersion) Command() string {
	return MsgVersionCmd
}

// MaxPayloadLength returns the maximum size of the MsgVersion message.
func (m *MsgVersion) MaxPayloadLength() uint32 {
	return 24
}
