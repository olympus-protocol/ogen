package bdb

import (
	"errors"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// ErrorBlockNodeSize returned when a blocknode size is above MaxBlockNodeSize
var ErrorBlockNodeSize = errors.New("blocknode size is too big")

// MaxBlockNodeSize is the maximum amount of bytes a BlockNodeDisk can be
const MaxBlockNodeSize = 624

// BlockNodeDisk is a block node stored on disk.
type BlockNodeDisk struct {
	StateRoot chainhash.Hash
	Height    uint64
	Slot      uint64
	Children  []chainhash.Hash
	Hash      chainhash.Hash
	Parent    chainhash.Hash
}

// Marshal encodes de data
func (bnd *BlockNodeDisk) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(bnd)
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
	return ssz.Unmarshal(d, bnd)
}
