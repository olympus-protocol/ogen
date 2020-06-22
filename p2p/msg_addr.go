package p2p

import (
	ssz "github.com/ferranbt/fastssz"
	"github.com/libp2p/go-libp2p-core/peer"
)

const MaxAddrPerMsg = 32
const MaxAddrPerPeer = 2

type MsgAddr struct {
	AddrList []peer.AddrInfo

	ssz.Marshaler
	ssz.Unmarshaler
}

// Marshal serializes the struct to bytes
func (m *MsgAddr) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the struct from bytes
func (m *MsgAddr) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

func (m *MsgAddr) Command() string {
	return MsgAddrCmd
}

func NewMsgAddr() *MsgAddr {
	return &MsgAddr{}
}
