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
	Votes:             &primitives.Votes{Votes: []*primitives.MultiValidatorVote{&MultiValidatorVote}},
	Txs:               &primitives.Txs{Txs: []*primitives.Tx{&TxSingle}},
	Deposits:          &primitives.Deposits{Deposits: []*primitives.Deposit{&Deposit}},
	Exits:             &primitives.Exits{Exits: []*primitives.Exit{&Exit}},
	VoteSlashings:     &primitives.VoteSlashings{VoteSlashings: []*primitives.VoteSlashing{&VoteSlashing}},
	RANDAOSlashings:   &primitives.RANDAOSlashings{RANDAOSlashings: []*primitives.RANDAOSlashing{&RANDAOSlashing}},
	ProposerSlashings: &primitives.ProposerSlashings{ProposerSlashings: []*primitives.ProposerSlashing{&ProposerSlashing}},
	GovernanceVotes:   &primitives.GovernanceVotes{GovernanceVotes: []*primitives.GovernanceVote{&GovernanceVote}},
	Signature:         sigB,
	RandaoSignature:   sigB,
}
