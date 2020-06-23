package p2p

const MaxAddrPerMsg = 32
const MaxAddrPerPeer = 2

type MsgAddr struct {
	AddrList [][]byte `ssz-size:"?,50" ssz-max:"16777216"`
}
