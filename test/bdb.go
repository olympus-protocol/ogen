package testdata

import (
	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var Hash, _ = chainhash.NewHashFromStr("8ce2ae922eea9be2772f620de7a4ae3aec61cce759f994502bfdfe5d9e6d52e0")

var BlockNode = bdb.BlockNodeDisk{
	StateRoot: *Hash,
	Height:    10298,
	Slot:      1000,
	Children:  [][32]byte{*Hash, *Hash, *Hash},
	Hash:      *Hash,
	Parent:    *Hash,
}
