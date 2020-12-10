package models

type BlockHeader struct {
	Hash                       string `gorm:"primarykey"`
	Version                    string
	Nonce                      string
	TxMerkleRoot               string
	TxMultiMerkleRoot          string
	VoteMerkleRoot             string
	DepositMerkleRoot          string
	ExitMerkleRoot             string
	VoteSlashingMerkleRoot     string
	RandaoSlashingMerkleRoot   string
	ProposerSlashingMerkleRoot string
	GovernanceVotesMarkleRoot  string
	PreviousBlockHash          string
	Timestamp                  string
	Slot                       string
	StateRoot                  string
	FeeAddress                 string
}
