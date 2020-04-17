package primitives

import (
	"bytes"
	"io"
	"time"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

var MaxBlockHeaderBytes = 76

type BlockHeader struct {
	Version       int32
	Nonce         int32
	MerkleRoot    chainhash.Hash
	PrevBlockHash chainhash.Hash
	Timestamp     time.Time
	Slot          uint64
}

func (bh *BlockHeader) Serialize(w io.Writer) error {
	sec := uint32(bh.Timestamp.Unix())
	err := serializer.WriteElements(w, bh.Version, bh.Nonce, bh.MerkleRoot, bh.PrevBlockHash, bh.Slot, sec)
	if err != nil {
		return err
	}
	return nil
}

func (bh *BlockHeader) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &bh.Version, &bh.Nonce, &bh.MerkleRoot, &bh.PrevBlockHash, &bh.Slot, (*serializer.Uint32Time)(&bh.Timestamp))
	if err != nil {
		return err
	}
	return nil
}

func (bh *BlockHeader) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = bh.Serialize(buf)
	return chainhash.DoubleHashH(buf.Bytes())
}
