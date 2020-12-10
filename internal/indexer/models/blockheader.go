package models

import "time"

type BlockHeader struct {
	Hash                       [32]byte `gorm:"primarykey"`
	Version                    [32]byte
	Nonce                      uint64
	TxMerkleRoot               [32]byte
	TxMultiMerkleRoot          [32]byte
	VoteMerkleRoot             [32]byte
	DepositMerkleRoot          [32]byte
	ExitMerkleRoot             [32]byte
	VoteSlashingMerkleRoot     [32]byte
	RandaoSlashingMerkleRoot   [32]byte
	ProposerSlashingMerkleRoot [32]byte
	GovernanceVotesMerkleRoot  [32]byte
	PreviousBlockHash          [32]byte
	Timestamp                  time.Time
	Slot                       uint64
	StateRoot                  [32]byte
	FeeAddress                 [20]byte
}
