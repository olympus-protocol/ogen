package bdb

import (
	ssz "github.com/ferranbt/fastssz"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// BlockNodeDisk is a block node stored on disk.
type BlockNodeDisk struct {
	StateRoot chainhash.Hash
	Height    uint64
	Slot      uint64
	Children  []chainhash.Hash
	Hash      chainhash.Hash
	Parent    chainhash.Hash

	ssz.Marshaler
	ssz.Unmarshaler
}

// Marshal serializes the struct to bytes
func (bnd *BlockNodeDisk) Marshal() ([]byte, error) {
	return bnd.MarshalSSZ()
}

// Unmarshal deserializes the struct from bytes
func (bnd *BlockNodeDisk) Unmarshal(b []byte) error {
	return bnd.UnmarshalSSZ(b)
}
