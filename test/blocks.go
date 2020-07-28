package testdata

import (
	"time"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var BlockRow = bdb.BlockNodeDisk{
	StateRoot: *Hash,
	Height:    1000,
	Slot:      1000,
	Children:  make([]chainhash.Hash, 32),
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
	Header:            BlockHeader,
	Votes:             []primitives.MultiValidatorVote{},
	Txs:               []primitives.Tx{},
	Deposits:          []primitives.Deposit{},
	Exits:             []primitives.Exit{},
	VoteSlashings:     []primitives.VoteSlashing{},
	RANDAOSlashings:   []primitives.RANDAOSlashing{},
	ProposerSlashings: []primitives.ProposerSlashing{},
	GovernanceVotes:   []primitives.GovernanceVote{},
	Signature:         sigB,
	RandaoSignature:   sigB,
}
