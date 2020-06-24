package p2p

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/prysmaticlabs/go-ssz"
)

const MaxAddrPerMsg = 32
const MaxAddrPerPeer = 2

type MsgAddr struct {
	AddrList []peer.AddrInfo
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
