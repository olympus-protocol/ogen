package p2p

import (
	"bytes"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
	"time"
)

var MaxBlockHeaderBytes = 76

type BlockHeader struct {
	Version       int32
	PrevBlockHash chainhash.Hash
	MerkleRoot    chainhash.Hash
	Timestamp     time.Time
}

func (bh *BlockHeader) Serialize(w io.Writer) error {
	sec := uint32(bh.Timestamp.Unix())
	err := serializer.WriteElements(w, bh.Version, bh.PrevBlockHash, bh.MerkleRoot, sec)
	if err != nil {
		return err
	}
	return nil
}

func (bh *BlockHeader) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &bh.Version, &bh.PrevBlockHash, &bh.MerkleRoot, (*serializer.Uint32Time)(&bh.Timestamp))
	if err != nil {
		return err
	}
	return nil
}

func (bh *BlockHeader) Hash() (chainhash.Hash, error) {
	buf := bytes.NewBuffer([]byte{})
	err := bh.Serialize(buf)
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.DoubleHashH(buf.Bytes()), nil
}

func NewBlockHeader(version int32, prevBlock chainhash.Hash, merkle chainhash.Hash, time time.Time) *BlockHeader {
	return &BlockHeader{
		Version:       version,
		PrevBlockHash: prevBlock,
		MerkleRoot:    merkle,
		Timestamp:     time,
	}
}
