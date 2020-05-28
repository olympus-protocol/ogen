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
	Version                    int32
	Nonce                      int32
	TxMerkleRoot               chainhash.Hash
	VoteMerkleRoot             chainhash.Hash
	DepositMerkleRoot          chainhash.Hash
	ExitMerkleRoot             chainhash.Hash
	VoteSlashingMerkleRoot     chainhash.Hash
	RANDAOSlashingMerkleRoot   chainhash.Hash
	ProposerSlashingMerkleRoot chainhash.Hash
	PrevBlockHash              chainhash.Hash
	Timestamp                  time.Time
	Slot                       uint64
	StateRoot                  chainhash.Hash
	FeeAddress                 [20]byte
}

// Serialize serializes the block header to the given writer.
func (bh *BlockHeader) Serialize(w io.Writer) error {
	sec := uint32(bh.Timestamp.Unix())
	err := serializer.WriteElements(w, bh.Version, bh.FeeAddress, bh.Nonce,
		bh.TxMerkleRoot, bh.VoteMerkleRoot, bh.PrevBlockHash,
		bh.Slot, bh.StateRoot, sec, bh.VoteSlashingMerkleRoot,
		bh.RANDAOSlashingMerkleRoot, bh.ProposerSlashingMerkleRoot)
	if err != nil {
		return err
	}
	return nil
}

// Deserialize deserializes the block header from the given reader.
func (bh *BlockHeader) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &bh.Version, &bh.FeeAddress, &bh.Nonce,
		&bh.TxMerkleRoot, &bh.VoteMerkleRoot, &bh.PrevBlockHash,
		&bh.Slot, &bh.StateRoot, (*serializer.Uint32Time)(&bh.Timestamp), &bh.VoteSlashingMerkleRoot,
		&bh.RANDAOSlashingMerkleRoot, &bh.ProposerSlashingMerkleRoot)
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
