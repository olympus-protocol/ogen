package primitives

import (
	"bytes"
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"github.com/prysmaticlabs/go-ssz"
)

var MaxBlockHeaderBytes = 76

type BlockHeader struct {
	Version                    uint32
	TxMerkleRoot               chainhash.Hash `ssz:"size=32"`
	VoteMerkleRoot             chainhash.Hash `ssz:"size=32"`
	DepositMerkleRoot          chainhash.Hash `ssz:"size=32"`
	ExitMerkleRoot             chainhash.Hash `ssz:"size=32"`
	VoteSlashingMerkleRoot     chainhash.Hash `ssz:"size=32"`
	RANDAOSlashingMerkleRoot   chainhash.Hash `ssz:"size=32"`
	ProposerSlashingMerkleRoot chainhash.Hash `ssz:"size=32"`
	GovernanceVotesMerkleRoot  chainhash.Hash `ssz:"size=32"`
	PrevBlockHash              chainhash.Hash `ssz:"size=32"`
	Timestamp                  uint64
	Slot                       uint64
	StateRoot                  chainhash.Hash `ssz:"size=32"`
	FeeAddress                 [20]byte       `ssz:"size=20"`
}

func (bh *BlockHeader) Marshal() ([]byte, error) {
	return ssz.Marshal(bh)
}

func (bh *BlockHeader) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, bh)
}

// Serialize serializes the block header to the given writer.
func (bh *BlockHeader) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, bh.Version, bh.FeeAddress,
		bh.TxMerkleRoot, bh.VoteMerkleRoot, bh.PrevBlockHash,
		bh.Slot, bh.StateRoot, bh.Timestamp, bh.VoteSlashingMerkleRoot,
		bh.RANDAOSlashingMerkleRoot, bh.ProposerSlashingMerkleRoot,
		bh.DepositMerkleRoot, bh.ExitMerkleRoot)
	if err != nil {
		return err
	}
	return nil
}

// Deserialize deserializes the block header from the given reader.
func (bh *BlockHeader) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &bh.Version, &bh.FeeAddress,
		&bh.TxMerkleRoot, &bh.VoteMerkleRoot, &bh.PrevBlockHash,
		&bh.Slot, &bh.StateRoot, &bh.Timestamp, &bh.VoteSlashingMerkleRoot,
		&bh.RANDAOSlashingMerkleRoot, &bh.ProposerSlashingMerkleRoot,
		&bh.DepositMerkleRoot, &bh.ExitMerkleRoot)
	if err != nil {
		return err
	}
	return nil
}

// Hash calculates the hash of the block header.
func (bh *BlockHeader) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = bh.Serialize(buf)
	return chainhash.DoubleHashH(buf.Bytes())
}
