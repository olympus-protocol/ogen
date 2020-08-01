package primitives

import (
	"errors"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// ErrorBlockHeaderSize returns when the Blockheader decompresed size is above MaxBlockHeaderBytes
var ErrorBlockHeaderSize = errors.New("blockheader size is too big")

// MaxBlockHeaderBytes is the maximum amount of bytes a header can contain.
const MaxBlockHeaderBytes = 372

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
func (b *BlockHeader) TxMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(b.TxMerkleRoot)
	return h
}

// VoteMerkleRootH returns the VoteMerkleRoot data as a hash struct
func (b *BlockHeader) VoteMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(b.VoteMerkleRoot)
	return h
}

// DepositMerkleRootH returns the DepositMerkleRoot data as a hash struct
func (b *BlockHeader) DepositMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(b.DepositMerkleRoot)
	return h
}

// ExitMerkleRootH returns the ExitMerkleRoot data as a hash struct
func (b *BlockHeader) ExitMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(b.ExitMerkleRoot)
	return h
}

// VoteSlashingMerkleRootH returns the VoteSlashingMerkleRoot data as a hash struct
func (b *BlockHeader) VoteSlashingMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(b.VoteSlashingMerkleRoot)
	return h
}

// GovernanceVotesMerkleRootH returns the GovernanceVotesMerkleRoot data as a hash struct
func (b *BlockHeader) GovernanceVotesMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(b.GovernanceVotesMerkleRoot)
	return h
}

// PrevBlockHashH returns the PrevBlockHash data as a hash struct
func (b *BlockHeader) PrevBlockHashH() *chainhash.Hash {
	h, _ := chainhash.NewHash(b.PrevBlockHash)
	return h
}

// RANDAOSlashingMerkleRootH returns the RANDAOSlashingMerkleRoot data as a hash struct
func (b *BlockHeader) RANDAOSlashingMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(b.RANDAOSlashingMerkleRoot)
	return h
}

// ProposerSlashingMerkleRootH returns the ProposerSlashingMerkleRoot data as a hash struct
func (b *BlockHeader) ProposerSlashingMerkleRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(b.ProposerSlashingMerkleRoot)
	return h
}

// StateRootH returns the StateRoot data as a hash struct
func (b *BlockHeader) StateRootH() *chainhash.Hash {
	h, _ := chainhash.NewHash(b.StateRoot)
	return h
}

// Marshal encodes the data.
func (b *BlockHeader) Marshal() ([]byte, error) {
	by, err := b.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(by) > MaxBlockHeaderBytes {
		return nil, ErrorBlockHeaderSize
	}
	return snappy.Encode(nil, by), nil
}

// Unmarshal decodes the data.
func (b *BlockHeader) Unmarshal(by []byte) error {
	d, err := snappy.Decode(nil, by)
	if err != nil {
		return err
	}
	if len(d) > MaxBlockHeaderBytes {
		return ErrorBlockHeaderSize
	}
	return b.UnmarshalSSZ(d)
}

// Hash calculates the hash of the block header.
func (b *BlockHeader) Hash() chainhash.Hash {
	by, _ := b.Marshal()
	return chainhash.HashH(by)
}
