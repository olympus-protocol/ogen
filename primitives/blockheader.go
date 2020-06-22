package primitives

import (
	"github.com/ferranbt/fastssz"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// BlockHeader is the primitive struct for a blockheader.
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

	ssz.Marshaler
	ssz.Unmarshaler
}

// Hash calculates the hash of the block header.
func (bh *BlockHeader) Hash() (chainhash.Hash, error) {
	b, err := bh.MarshalSSZ()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.DoubleHashH(b), nil
}
