package bdb

import (
	"errors"

	"github.com/golang/snappy"
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
func (bnd *BlockNodeDisk) Marshal() ([]byte, error) {
	b, err := bnd.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxBlockNodeSize {
		return nil, ErrorBlockNodeSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data
func (bnd *BlockNodeDisk) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxBlockNodeSize {
		return ErrorBlockNodeSize
	}
	return bnd.UnmarshalSSZ(d)
}
