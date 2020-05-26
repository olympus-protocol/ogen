package bdb

import (
	"bytes"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// BlockNodeDisk is a block node stored on disk.
type BlockNodeDisk struct {
	StateRoot chainhash.Hash
	Height    uint64
	Slot      uint64
	Children  []chainhash.Hash
	Hash      chainhash.Hash
	Parent    chainhash.Hash
}

// Serialize serializes a block node disk to bytes.
func (bnd *BlockNodeDisk) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	err := serializer.WriteVarInt(buf, uint64(len(bnd.Children)))
	if err != nil {
		return nil, err
	}

	for _, c := range bnd.Children {
		if err := serializer.WriteElement(buf, c); err != nil {
			return nil, err
		}
	}

	err = serializer.WriteElements(buf, bnd.StateRoot, bnd.Height, bnd.Hash, bnd.Parent, bnd.Slot)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Deserialize deserializes a block node disk from bytes.
func (bnd *BlockNodeDisk) Deserialize(b []byte) error {
	buf := bytes.NewBuffer(b)

	numChildren, err := serializer.ReadVarInt(buf)
	if err != nil {
		return err
	}

	bnd.Children = make([]chainhash.Hash, numChildren)
	for i := range bnd.Children {
		if err := serializer.ReadElement(buf, &bnd.Children[i]); err != nil {
			return err
		}
	}

	err = serializer.ReadElements(buf, &bnd.StateRoot, &bnd.Height, &bnd.Hash, &bnd.Parent, &bnd.Slot)
	if err != nil {
		return err
	}

	return nil
}
