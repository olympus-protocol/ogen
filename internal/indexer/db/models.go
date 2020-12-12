package db

import "time"

type Block struct {
	Hash     [32]byte `gorm:"primaryKey"`
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
	Account [20]byte `gorm:"primaryKey"`
	Balance uint64
	Nonce   uint64
}

type BlockHeader struct {
	Hash                       [32]byte `gorm:"primaryKey"`
	Version                    uint64
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

type CoinProofs struct {
	Hash          [32]byte `gorm:"primarykey"`
	Transaction   [192]byte
	RedeemAccount [44]byte
}

type Deposit struct {
	Hash      [32]byte `gorm:"primaryKey"`
	BlockHash [32]byte
	PublicKey [48]byte
	Data      DepositData `gorm:"foreignKey:Hash"`
}

type DepositData struct {
	Hash              [32]byte
	PublicKey         [48]byte `gorm:"primaryKey"`
	ProofOfPossession [96]byte
	WithdrawalAddress [20]byte
}

type Epoch struct {
	Epoch                   uint64 `gorm:"primaryKey"`
	Slot1                   Slot   `gorm:"foreignKey:Slot"`
	Slot2                   Slot   `gorm:"foreignKey:Slot"`
	Slot3                   Slot   `gorm:"foreignKey:Slot"`
	Slot4                   Slot   `gorm:"foreignKey:Slot"`
	Slot5                   Slot   `gorm:"foreignKey:Slot"`
	ParticipationPercentage []byte
	Finalized               bool
	Justified               bool
	Randao                  [32]byte
}

type Exit struct {
	Hash                [32]byte `gorm:"primarykey"`
	BlockHash           [32]byte
	ValidatorPublicKey  [48]byte
	WithdrawalPublicKey [48]byte
}

type PartialExit struct {
	Hash                [32]byte
	BlockHash           [32]byte
	ValidatorPublicKey  [48]byte
	WithdrawalPublicKey [48]byte
	Amount              uint64
}

type Slot struct {
	Slot          uint64 `gorm:"primaryKey"`
	BlockHash     [32]byte
	ProposerIndex uint64
	Proposed      bool
}

type Tx struct {
	BlockHash         [32]byte
	Hash              [32]byte `gorm:"primarykey"`
	ToAddress         [20]byte
	FromPublicKeyHash [20]byte
	FromPublicKey     [48]byte
	Amount            uint64
	Nonce             uint64
	Fee               uint64
}

type Validator struct {
	Balance          uint64
	PubKey           [48]byte `gorm:"primaryKey"`
	PayeeAddress     [20]byte
	Status           uint64
	FirstActiveEpoch uint64
	LastActiveEpoch  uint64
}

type Vote struct {
	BlockHash             [32]byte
	ParticipationBitfield []byte
	Hash                  [32]byte `gorm:"primaryKey"`
	Data                  VoteData `gorm:"foreignKey:Hash"`
}

type VoteData struct {
	Hash            [32]byte `gorm:"primaryKey"`
	Slot            uint64
	FromEpoch       uint64
	FromHash        [32]byte
	ToEpoch         uint64
	ToHash          [32]byte
	BeaconBlockHash [32]byte
	Nonce           uint64
}

type State struct {
	Key string `gorm:"primaryKey"`
	Raw []byte
}
