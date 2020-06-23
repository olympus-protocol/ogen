package bdb

// BlockNodeDisk is a block node stored on disk.
type BlockNodeDisk struct {
	StateRoot []byte `ssz-size:"32"`
	Height    uint64
	Slot      uint64
	Children  [][]byte `ssz-size:"?,32" ssz-max:"16777216"`
	Hash      []byte   `ssz-size:"32"`
	Parent    []byte   `ssz-size:"32"`
}
