package msg

import (
	"github.com/libp2p/go-libp2p-core/peer"
)

const MaxAddrPerMsg = 32
const MaxAddrPerPeer = 2

type MsgAddr struct {
	AddrList [][]byte `ssz-size:"?,50" ssz-max:"16777216"`
}

func (m *MsgAddr) Command() string {
	return MsgAddrCmd
}
