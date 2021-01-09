package primitives

import "github.com/golang/snappy"

// BlockNodeDisk is a block node stored on disk.
type BlockNodeDisk struct {
	Height    uint64
	Slot      uint64
	Children  [][32]byte `ssz-max:"64"`
	Hash      [32]byte   `ssz-size:"32"`
	Parent    [32]byte   `ssz-size:"32"`
}

// Marshal encodes de data
func (b *BlockNodeDisk) Marshal() ([]byte, error) {
	ser, err := b.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, ser), nil
}

// Unmarshal decodes the data
func (b *BlockNodeDisk) Unmarshal(by []byte) error {
	des, err := snappy.Decode(nil, by)
	if err != nil {
		return err
	}
	return b.UnmarshalSSZ(des)
}
