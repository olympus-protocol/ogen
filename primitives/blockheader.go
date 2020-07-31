package primitives

import (
	"errors"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// ErrorBlockHeaderSize returns when the Blockheader decompresed size is above MaxBlockHeaderBytes
var ErrorBlockHeaderSize = errors.New("blockheader size is too big")

// MaxBlockHeaderBytes is the maximum amount of bytes a header can contain.
const MaxBlockHeaderBytes = 376

// BlockHeader is the container of merkle roots for the blockchain
type BlockHeader struct {
	Version                    uint64
	Nonce                      uint64
	TxMerkleRoot               [32]byte `ssz-size:"32"`
	VoteMerkleRoot             [32]byte `ssz-size:"32"`
	DepositMerkleRoot          [32]byte `ssz-size:"32"`
	ExitMerkleRoot             [32]byte `ssz-size:"32"`
	VoteSlashingMerkleRoot     [32]byte `ssz-size:"32"`
	RANDAOSlashingMerkleRoot   [32]byte `ssz-size:"32"`
	ProposerSlashingMerkleRoot [32]byte `ssz-size:"32"`
	GovernanceVotesMerkleRoot  [32]byte `ssz-size:"32"`
	PrevBlockHash              [32]byte `ssz-size:"32"`
	Timestamp                  uint64
	Slot                       uint64
	StateRoot                  [32]byte `ssz-size:"32"`
	FeeAddress                 [20]byte `ssz-size:"20"`
}

// TxMerkleRootH returns the TxMerkleRoot data as a hash struct
func (bh *BlockHeader) TxMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(bh.TxMerkleRoot)
	return h
}

// VoteMerkleRootH returns the VoteMerkleRoot data as a hash struct
func (bh *BlockHeader) VoteMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(bh.VoteMerkleRoot)
	return h
}

// DepositMerkleRootH returns the DepositMerkleRoot data as a hash struct
func (bh *BlockHeader) DepositMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(bh.DepositMerkleRoot)
	return h
}

// ExitMerkleRootH returns the ExitMerkleRoot data as a hash struct
func (bh *BlockHeader) ExitMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(bh.ExitMerkleRoot)
	return h
}

// VoteSlashingMerkleRootH returns the VoteSlashingMerkleRoot data as a hash struct
func (bh *BlockHeader) VoteSlashingMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(bh.VoteSlashingMerkleRoot)
	return h
}

// GovernanceVotesMerkleRootH returns the GovernanceVotesMerkleRoot data as a hash struct
func (bh *BlockHeader) GovernanceVotesMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(bh.GovernanceVotesMerkleRoot)
	return h
}

// PrevBlockHashH returns the PrevBlockHash data as a hash struct
func (bh *BlockHeader) PrevBlockHashH() *chainhash.Hash {
	h, _ := chainhash.NewHash(bh.PrevBlockHash)
	return h
}

// RANDAOSlashingMerkleRootH returns the RANDAOSlashingMerkleRoot data as a hash struct
func (bh *BlockHeader) RANDAOSlashingMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(bh.RANDAOSlashingMerkleRoot)
	return h
}

// ProposerSlashingMerkleRootH returns the ProposerSlashingMerkleRoot data as a hash struct
func (bh *BlockHeader) ProposerSlashingMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(bh.ProposerSlashingMerkleRoot)
	return h
}

// StateRootH returns the StateRoot data as a hash struct
func (bh *BlockHeader) StateRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(bh.StateRoot)
	return h
}

// Marshal encodes the data.
func (bh *BlockHeader) Marshal() ([]byte, error) {
	b, err := bh.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxBlockHeaderBytes {
		return nil, ErrorBlockHeaderSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (bh *BlockHeader) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxBlockHeaderBytes {
		return ErrorBlockHeaderSize
	}
	return bh.UnmarshalSSZ(d)
}

// Hash calculates the hash of the block header.
func (bh *BlockHeader) Hash() chainhash.Hash {
	b, _ := bh.Marshal()
	return chainhash.HashH(b)
}
