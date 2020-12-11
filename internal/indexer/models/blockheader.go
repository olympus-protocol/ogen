package models

import "time"

type BlockHeader struct {
	Hash                       string `gorm:"primaryKey"`
	Version                    uint64
	Nonce                      string
	TxMerkleRoot               string
	TxMultiMerkleRoot          string
	VoteMerkleRoot             string
	DepositMerkleRoot          string
	ExitMerkleRoot             string
	VoteSlashingMerkleRoot     string
	RandaoSlashingMerkleRoot   string
	ProposerSlashingMerkleRoot string
	GovernanceVotesMerkleRoot  string
	PreviousBlockHash          string
	Timestamp                  time.Time
	Slot                       uint64
	StateRoot                  string
	FeeAddress                 string
}
