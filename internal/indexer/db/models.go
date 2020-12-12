package db

import (
	"time"
)

type Block struct {
	Hash     []byte `gorm:"primaryKey"`
	Height   uint64
	Slot     uint64
	Header   BlockHeader `gorm:"foreignKey:Hash"`
	Txs      []Tx
	Deposits []Deposit
	Votes    []Vote
	Exits    []Exit
	RawBlock []byte
}

type Account struct {
	Account string `gorm:"primaryKey"`
	Balance uint64
	Nonce   uint64
}

type BlockHeader struct {
	Hash                       []byte `gorm:"primaryKey"`
	Version                    uint64
	Nonce                      []byte
	TxMerkleRoot               []byte
	TxMultiMerkleRoot          []byte
	VoteMerkleRoot             []byte
	DepositMerkleRoot          []byte
	ExitMerkleRoot             []byte
	VoteSlashingMerkleRoot     []byte
	RandaoSlashingMerkleRoot   []byte
	ProposerSlashingMerkleRoot []byte
	GovernanceVotesMerkleRoot  []byte
	PreviousBlockHash          []byte
	Timestamp                  time.Time
	Slot                       uint64
	StateRoot                  []byte
	FeeAddress                 []byte
}

type CoinProofs struct {
	Hash          []byte `gorm:"primarykey"`
	Transaction   []byte
	RedeemAccount []byte
}

type Deposit struct {
	Hash      []byte `gorm:"primaryKey"`
	BlockHash []byte
	PublicKey []byte
	Data      DepositData `gorm:"foreignKey:Hash"`
}

type DepositData struct {
	Hash              []byte
	PublicKey         []byte `gorm:"primaryKey"`
	ProofOfPossession []byte
	WithdrawalAddress []byte
}

type Epoch struct {
	Epoch                   uint64 `gorm:"primaryKey"`
	Slot1                   uint64
	Slot2                   uint64
	Slot3                   uint64
	Slot4                   uint64
	Slot5                   uint64
	ParticipationPercentage []byte
	Finalized               bool
	Justified               bool
	Randao                  []byte
}

type Exit struct {
	Hash                []byte `gorm:"primarykey"`
	BlockHash           []byte
	ValidatorPublicKey  []byte
	WithdrawalPublicKey []byte
}

type PartialExit struct {
	Hash                []byte
	BlockHash           []byte
	ValidatorPublicKey  []byte
	WithdrawalPublicKey []byte
	Amount              uint64
}

type Slot struct {
	Slot          uint64 `gorm:"primaryKey"`
	BlockHash     []byte
	ProposerIndex uint64
	Proposed      bool
}

type Tx struct {
	BlockHash         []byte
	Hash              []byte `gorm:"primarykey"`
	ToAddress         []byte
	FromPublicKeyHash []byte
	FromPublicKey     []byte
	Amount            uint64
	Nonce             uint64
	Fee               uint64
}

type Validator struct {
	Balance          uint64
	PubKey           []byte `gorm:"primaryKey"`
	PayeeAddress     []byte
	Status           uint64
	FirstActiveEpoch uint64
	LastActiveEpoch  uint64
}

type Vote struct {
	BlockHash             []byte
	ParticipationBitfield []byte
	Hash                  []byte   `gorm:"primaryKey"`
	Data                  VoteData `gorm:"foreignKey:Hash"`
}

type VoteData struct {
	Hash            []byte `gorm:"primaryKey"`
	Slot            uint64
	FromEpoch       uint64
	FromHash        []byte
	ToEpoch         uint64
	ToHash          []byte
	BeaconBlockHash []byte
	Nonce           []byte
}

type State struct {
	Key             string `gorm:"primaryKey"`
	Raw             []byte
	LastBlock       []byte
	LastBlockHeight uint64
}
