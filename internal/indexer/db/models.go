package db

import "time"

type Block struct {
	Hash     string `gorm:"primaryKey"`
	Height   uint64
	Slot     uint64
	Header   BlockHeader `gorm:"foreignKey:Hash"`
	Txs      []Tx
	Deposits []Deposit
	Votes    []Vote
	Exits    []Exit
}

type Account struct {
	Account string `gorm:"primaryKey"`
	Balance uint64
	Nonce   uint64
}

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

type CoinProofs struct {
	Hash          string `gorm:"primarykey"`
	Transaction   string
	RedeemAccount string
}

type Deposit struct {
	Hash      string `gorm:"primaryKey"`
	BlockHash string
	PublicKey string
	Data      DepositData `gorm:"foreignKey:Hash"`
}

type DepositData struct {
	Hash              string
	PublicKey         string `gorm:"primaryKey"`
	ProofOfPossession string
	WithdrawalAddress string
}

type Epoch struct {
	Epoch                   uint64 `gorm:"primaryKey"`
	Slot1                   Slot   `gorm:"foreignKey:Slot"`
	Slot2                   Slot   `gorm:"foreignKey:Slot"`
	Slot3                   Slot   `gorm:"foreignKey:Slot"`
	Slot4                   Slot   `gorm:"foreignKey:Slot"`
	Slot5                   Slot   `gorm:"foreignKey:Slot"`
	ParticipationPercentage uint64
	Finalized               bool
	Justified               bool
	Randao                  string
}

type Exit struct {
	Hash                string `gorm:"primarykey"`
	BlockHash           string
	ValidatorPublicKey  string
	WithdrawalPublicKey string
}

type PartialExit struct {
	Hash                string
	BlockHash           string
	ValidatorPublicKey  string
	WithdrawalPublicKey string
	Amount              uint64
}

type Slot struct {
	Slot          uint64 `gorm:"primaryKey"`
	BlockHash     string
	ProposerIndex uint64
	Proposed      bool
}

type Tx struct {
	BlockHash         string
	Hash              string `gorm:"primarykey"`
	ToAddress         string
	FromPublicKeyHash string
	FromPublicKey     string
	Amount            uint64
	Nonce             uint64
	Fee               uint64
}

type Validator struct {
	Balance          uint64
	PubKey           string `gorm:"primaryKey"`
	PayeeAddress     string
	Status           uint64
	FirstActiveEpoch uint64
	LastActiveEpoch  uint64
}

type Vote struct {
	BlockHash             string
	ParticipationBitfield string
	Hash                  string   `gorm:"primaryKey"`
	Data                  VoteData `gorm:"foreignKey:Hash"`
}

type VoteData struct {
	Hash            string `gorm:"primaryKey"`
	Slot            uint64
	FromEpoch       uint64
	FromHash        string
	ToEpoch         uint64
	ToHash          string
	BeaconBlockHash string
	Nonce           string
}
