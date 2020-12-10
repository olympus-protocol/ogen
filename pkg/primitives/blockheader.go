package primitives

import (
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// BlockHeader is the container of merkle roots for the blockchain
type BlockHeader struct {
	Version                    uint64
	Nonce                      uint64
	TxMerkleRoot               [32]byte `ssz-size:"32"`
	TxMultiMerkleRoot          [32]byte `ssz-size:"32"`
	VoteMerkleRoot             [32]byte `ssz-size:"32"`
	DepositMerkleRoot          [32]byte `ssz-size:"32"`
	ExitMerkleRoot             [32]byte `ssz-size:"32"`
	VoteSlashingMerkleRoot     [32]byte `ssz-size:"32"`
	RANDAOSlashingMerkleRoot   [32]byte `ssz-size:"32"`
	ProposerSlashingMerkleRoot [32]byte `ssz-size:"32"`
	GovernanceVotesMerkleRoot  [32]byte `ssz-size:"32"`
	CoinProofsMerkleRoot       [32]byte `ssz-size:"32"`
	PrevBlockHash              [32]byte `ssz-size:"32"`
	Timestamp                  uint64
	Slot                       uint64
	StateRoot                  [32]byte `ssz-size:"32"`
	FeeAddress                 [20]byte `ssz-size:"20"`
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
