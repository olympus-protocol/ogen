package primitives

import (
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// BlockHeader is the container of merkle roots for the blockchain
type BlockHeader struct {
	Version   uint64
	Nonce     uint64
	Timestamp uint64
	Slot      uint64

	TxMerkleRoot               [32]byte
	TxMultiMerkleRoot          [32]byte
	VoteMerkleRoot             [32]byte
	DepositMerkleRoot          [32]byte
	ExitMerkleRoot             [32]byte
	PartialExitMerkleRoot      [32]byte
	VoteSlashingMerkleRoot     [32]byte
	RANDAOSlashingMerkleRoot   [32]byte
	ProposerSlashingMerkleRoot [32]byte
	GovernanceVotesMerkleRoot  [32]byte
	CoinProofsMerkleRoot       [32]byte
	ExecutionsMerkleRoot       [32]byte
	PrevBlockHash              [32]byte
	StateRoot                  [32]byte

	FeeAddress [20]byte
}

// Marshal encodes the data.
func (b *BlockHeader) Marshal() ([]byte, error) {
	return b.MarshalSSZ()
}

// Unmarshal decodes the data.
func (b *BlockHeader) Unmarshal(by []byte) error {
	return b.UnmarshalSSZ(by)
}

// Hash calculates the hash of the block header.
func (b *BlockHeader) Hash() chainhash.Hash {
	by, _ := b.Marshal()
	return chainhash.HashH(by)
}
