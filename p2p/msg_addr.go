package p2p

import (
	"github.com/libp2p/go-libp2p-core/peer"
)

const MaxAddrPerMsg = 32
const MaxAddrPerPeer = 2

type MsgAddr struct {
	AddrList []peer.AddrInfo
}

// Marshal serializes the struct to bytes
func (m *MsgAddr) Marshal() ([]byte, error) {
	return m.Marshal()
}

// Unmarshal deserializes the struct from bytes
func (m *MsgAddr) Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *MsgAddr) Command() string {
	return MsgAddrCmd
}

func NewMsgAddr() *MsgAddr {
	return &MsgAddr{}
}
