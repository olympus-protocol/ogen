package blockdb

import (
	"errors"
)

// ErrorBlockNodeSize returned when a blocknode size is above MaxBlockNodeSize
var ErrorBlockNodeSize = errors.New("blocknode size is too big")

// MaxBlockNodeSize is the maximum amount of bytes a BlockNodeDisk can be
const MaxBlockNodeSize = 2160

// BlockNodeDisk is a block node stored on disk.
type BlockNodeDisk struct {
	StateRoot [32]byte `ssz-size:"32"`
	Height    uint64
	Slot      uint64
	Children  [][32]byte `ssz-max:"64"`
	Hash      [32]byte   `ssz-size:"32"`
	Parent    [32]byte   `ssz-size:"32"`
}

// Marshal encodes de data
func (b *BlockNodeDisk) Marshal() ([]byte, error) {
	return b.MarshalSSZ()
}

// Unmarshal decodes the data
func (b *BlockNodeDisk) Unmarshal(by []byte) error {
	return b.UnmarshalSSZ(by)
}
