package testdata

import (
	"time"

	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
)

var Addr = [500]byte{}

func init() {
	addrB := []byte("/ip4/0.0.0.0/tcp/24126/p2p/12D3KooWCNzTeoaaVGKGRqS7vfzFxZMvXccNRCabCFQxS5cQaf6F")
	copy(Addr[:], addrB)
}

var Header = p2p.MessageHeader{
	Magic:    99999999,
	Command:  [40]byte{0x67, 0x65, 0x74, 0x62, 0x6c, 0x6f, 0x63, 0x6b},
	Length:   123123123,
	Checksum: [4]byte{0x0, 0x0, 0x0, 0x0},
}

var MsgGetAddr = p2p.MsgGetAddr{}

var MsgAddr = p2p.MsgAddr{
	Addr: [32][500]byte{Addr, Addr},
}

var MsgBlocks = p2p.MsgBlocks{
	Blocks: []*primitives.Block{&Block, &Block, &Block, &Block, &Block},
}

var MsgGetBlocks = p2p.MsgGetBlocks{
	HashStop:      *Hash,
	LocatorHashes: [64][32]byte{*Hash, *Hash, *Hash},
}

var MsgVersion = p2p.MsgVersion{
	LastBlock: 90000,
	Nonce:     123123123,
	Timestamp: uint64(time.Now().Unix()),
}
