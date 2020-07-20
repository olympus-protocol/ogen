package primitives

import (
	"github.com/golang/snappy"
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
	b, err := ssz.Marshal(bh)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (bh *BlockHeader) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(d, bh)
}

// Hash calculates the hash of the block header.
func (bh *BlockHeader) Hash() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(bh)
	return chainhash.Hash(hash)
}
