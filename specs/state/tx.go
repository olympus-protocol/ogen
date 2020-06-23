package state

type TxLocator struct {
	Hash  []byte `ssz-size:"32"`
	Block []byte `ssz-size:"32"`
	Index uint32
}

type Transfer struct {
	To            []byte `ssz-size:"20"`
	FromPublicKey []byte `ssz-size:"48"`
	Amount        uint64
	Nonce         uint64
	Fee           uint64
	Signature     []byte `ssz-size:"96"`
}

type Tx struct {
	Version uint32
	Type    uint32
	Payload *Transfer
}
