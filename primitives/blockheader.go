package primitives

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

var MaxBlockHeaderBytes = 76

type BlockHeader struct {
	Version                    uint32
	Nonce                      uint32
	TxMerkleRoot               chainhash.Hash
	VoteMerkleRoot             chainhash.Hash
	DepositMerkleRoot          chainhash.Hash
	ExitMerkleRoot             chainhash.Hash
	VoteSlashingMerkleRoot     chainhash.Hash
	RANDAOSlashingMerkleRoot   chainhash.Hash
	ProposerSlashingMerkleRoot chainhash.Hash
	GovernanceVotesMerkleRoot  chainhash.Hash
	PrevBlockHash              chainhash.Hash
	Timestamp                  uint64
	Slot                       uint64
	StateRoot                  chainhash.Hash
	FeeAddress                 [20]byte
}

// Marshal encodes the data.
func (bh *BlockHeader) Marshal() ([]byte, error) {
	return ssz.Marshal(bh)
}

// Unmarshal decodes the data.
func (bh *BlockHeader) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, bh)
}

// Hash calculates the hash of the block header.
func (bh *BlockHeader) Hash() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(bh)
	return chainhash.Hash(hash)
}
