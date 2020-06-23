package primitives

import "github.com/olympus-protocol/ogen/utils/chainhash"

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

func (c *Transfer) Hash() chainhash.Hash {
	// TODO handle error
	b, _ := c.MarshalSSZ()
	return chainhash.HashH(b)
}

type Tx struct {
	Version uint32
	Type    uint32
	Payload *Transfer
}

// Hash calculates the transaction hash.
func (t *Tx) Hash() chainhash.Hash {
	// TODO handle error
	b, _ := t.MarshalSSZ()
	return chainhash.DoubleHashH(b)
}
