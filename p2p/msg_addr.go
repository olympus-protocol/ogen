package p2p

import (
	"github.com/golang/snappy"
	"github.com/prysmaticlabs/go-ssz"
)

const MaxAddrPerMsg = 500
const MaxAddrPerPeer = 2

type MsgAddr struct {
	Addr [][]byte
}

// Marshal serializes the data to bytes
func (m *MsgAddr) Marshal() ([]byte, error) {
	return ssz.Marshal(m)
}

// Unmarshal deserializes the data
func (m *MsgAddr) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(d, m)
}

func (m *MsgAddr) Command() string {
	return MsgAddrCmd
}

func (m *MsgAddr) MaxPayloadLength() uint32 {
	netAddressSize := 26 // Max NetAddress size
	return uint32(MaxAddrPerMsg * netAddressSize)
}

func NewMsgAddr() *MsgAddr {
	return &MsgAddr{}
}
