package testdata

import (
	"time"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/primitives"
)

var BlockRow = bdb.BlockNodeDisk{
	StateRoot: *Hash,
	Height:    1000,
	Slot:      1000,
	Children:  make([][32]byte, 32),
	Hash:      *Hash,
	Parent:    *Hash,
}

var BlockHeader = primitives.BlockHeader{
	Version:                    1,
	Nonce:                      123123,
	TxMerkleRoot:               *Hash,
	VoteMerkleRoot:             *Hash,
	DepositMerkleRoot:          *Hash,
	ExitMerkleRoot:             *Hash,
	VoteSlashingMerkleRoot:     *Hash,
	RANDAOSlashingMerkleRoot:   *Hash,
	ProposerSlashingMerkleRoot: *Hash,
	GovernanceVotesMerkleRoot:  *Hash,
	PrevBlockHash:              *Hash,
	Timestamp:                  uint64(time.Now().Unix()),
	Slot:                       123,
	StateRoot:                  *Hash,
	FeeAddress:                 [20]byte{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
}

var Block = primitives.Block{
	Header:            &BlockHeader,
	Votes:             &primitives.Votes{},
	Txs:               &primitives.Txs{},
	Deposits:          &primitives.Deposits{},
	Exits:             &primitives.Exits{},
	VoteSlashings:     &primitives.VoteSlashings{},
	RANDAOSlashings:   &primitives.RANDAOSlashings{},
	ProposerSlashings: &primitives.ProposerSlashings{},
	GovernanceVotes:   &primitives.GovernanceVotes{},
	Signature:         sigB,
	RandaoSignature:   sigB,
}
