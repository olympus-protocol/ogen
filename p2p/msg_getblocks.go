package p2p

type MsgGetBlocks struct {
	HashStop      []byte   `ssz-size:"32"`
	LocatorHashes [][]byte `ssz-size:"?,32" ssz-max:"16777216"`
}
