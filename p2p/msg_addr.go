package p2p

import (
	"github.com/golang/snappy"
)

// MaxAddrPerMsg defines the maximum address that can be added into an addr message.
const MaxAddrPerMsg = 32

// MsgAddr is the struct for the response of getaddr.
type MsgAddr struct {
	Addr [][64]byte `ssz-max:"32"`
}

// Marshal serializes the data to bytes
func (m *MsgAddr) Marshal() ([]byte, error) {
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
func (m *MsgAddr) Unmarshal(b []byte) error {
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
func (m *MsgAddr) Command() string {
	return MsgAddrCmd
}

// MaxPayloadLength returns the maximum size of the MsgAddr message.
func (m *MsgAddr) MaxPayloadLength() uint64 {
	return uint64(MaxAddrPerMsg*64) + 4
}
