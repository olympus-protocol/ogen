package p2p

import (
	"github.com/golang/snappy"
	"github.com/prysmaticlabs/go-ssz"
)

// MaxAddrPerMsg defines the maximum address that can be added into an addr message.
const MaxAddrPerMsg = 32

// MaxAddrPerPeer defines the maximum amount of address that a single peer can send.
const MaxAddrPerPeer = 2

// MsgAddr is the struct for the response of getaddr.
type MsgAddr struct {
	Addr [][]byte
}

// Marshal serializes the data to bytes
func (m *MsgAddr) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(m)
	if err != nil {
		return nil, err
	}
	if uint32(len(b)) > m.MaxPayloadLength() {
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
	if uint32(len(d)) > m.MaxPayloadLength() {
		return ErrorSizeExceed
	}
	return ssz.Unmarshal(d, m)
}

// Command returns the message topic
func (m *MsgAddr) Command() string {
	return MsgAddrCmd
}

// MaxPayloadLength returns the maximum size of the MsgAddr message.
func (m *MsgAddr) MaxPayloadLength() uint32 {
	netAddressSize := 500 // There is no a specific maximum size for ma formatted address.
	return uint32(MaxAddrPerMsg * netAddressSize)
}
