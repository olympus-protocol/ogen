package p2p

import (
	"github.com/prysmaticlabs/go-ssz"
)

const MaxAddrPerMsg = 500
const MaxAddrPerPeer = 2

type MsgAddr struct {
	Addr [][]byte
}

func (m *MsgAddr) Marshal() ([]byte, error) {
	return ssz.Marshal(m)
}

func (m *MsgAddr) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, m)
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
