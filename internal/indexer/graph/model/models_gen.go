// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type Account struct {
	Account string `json:"Account"`
	Balance int    `json:"Balance"`
	Nonce   int    `json:"Nonce"`
}

type Block struct {
	Hash     string       `json:"Hash"`
	Height   int          `json:"Height"`
	Slot     int          `json:"Slot"`
	Header   *BlockHeader `json:"Header"`
	Txs      []*Tx        `json:"Txs"`
	Deposits []*Deposit   `json:"Deposits"`
	Votes    []*Vote      `json:"Votes"`
	Exits    []*Exit      `json:"Exits"`
	RawBlock string       `json:"RawBlock"`
}

type BlockHeader struct {
	Hash                       string `json:"Hash"`
	Version                    int    `json:"Version"`
	Nonce                      string `json:"Nonce"`
	TxMerkleRoot               string `json:"TxMerkleRoot"`
	TxMultiMerkleRoot          string `json:"TxMultiMerkleRoot"`
	VoteMerkleRoot             string `json:"VoteMerkleRoot"`
	DepositMerkleRoot          string `json:"DepositMerkleRoot"`
	ExitMerkleRoot             string `json:"ExitMerkleRoot"`
	VoteSlashingMerkleRoot     string `json:"VoteSlashingMerkleRoot"`
	RandaoSlashingMerkleRoot   string `json:"RandaoSlashingMerkleRoot"`
	ProposerSlashingMerkleRoot string `json:"ProposerSlashingMerkleRoot"`
	GovernanceVotesMerkleRoot  string `json:"GovernanceVotesMerkleRoot"`
	PreviousBlockHash          string `json:"PreviousBlockHash"`
	Timestamp                  string `json:"Timestamp"`
	Slot                       int    `json:"Slot"`
	StateRoot                  string `json:"StateRoot"`
	FeeAddress                 string `json:"FeeAddress"`
}

type CoinProofs struct {
	Hash          string `json:"Hash"`
	Transaction   string `json:"Transaction"`
	RedeemAccount string `json:"RedeemAccount"`
}

type Deposit struct {
	Hash      string       `json:"Hash"`
	BlockHash string       `json:"BlockHash"`
	PublicKey string       `json:"PublicKey"`
	Data      *DepositData `json:"Data"`
}

type DepositData struct {
	Hash              string `json:"Hash"`
	PublicKey         string `json:"PublicKey"`
	ProofOfPossession string `json:"ProofOfPossession"`
	WithdrawalAddress string `json:"WithdrawalAddress"`
}

type Epoch struct {
	Epoch                   int    `json:"Epoch"`
	Slot1                   int    `json:"Slot1"`
	Slot2                   int    `json:"Slot2"`
	Slot3                   int    `json:"Slot3"`
	Slot4                   int    `json:"Slot4"`
	Slot5                   int    `json:"Slot5"`
	ParticipationPercentage string `json:"ParticipationPercentage"`
	Finalized               bool   `json:"Finalized"`
	Justified               bool   `json:"Justified"`
	Randao                  string `json:"Randao"`
}

type Exit struct {
	Hash                string `json:"Hash"`
	BlockHash           string `json:"BlockHash"`
	ValidatorPublicKey  string `json:"ValidatorPublicKey"`
	WithdrawalPublicKey string `json:"WithdrawalPublicKey"`
}

type PartialExit struct {
	Hash                string `json:"Hash"`
	BlockHash           string `json:"BlockHash"`
	ValidatorPublicKey  string `json:"ValidatorPublicKey"`
	WithdrawalPublicKey string `json:"WithdrawalPublicKey"`
	Amount              int    `json:"Amount"`
}

type Tx struct {
	BlockHash         string `json:"BlockHash"`
	Hash              string `json:"Hash"`
	ToAddress         string `json:"ToAddress"`
	FromPublicKeyHash string `json:"FromPublicKeyHash"`
	FromPublicKey     string `json:"FromPublicKey"`
	Amount            int    `json:"Amount"`
	Nonce             int    `json:"Nonce"`
	Fee               int    `json:"Fee"`
}

type Validator struct {
	Balance          int    `json:"Balance"`
	PubKey           string `json:"PubKey"`
	PayeeAddress     string `json:"PayeeAddress"`
	Status           int    `json:"Status"`
	FirstActiveEpoch int    `json:"FirstActiveEpoch"`
	LastActiveEpoch  int    `json:"LastActiveEpoch"`
}

type Vote struct {
	BlockHash             string    `json:"BlockHash"`
	ParticipationBitfield string    `json:"ParticipationBitfield"`
	Hash                  string    `json:"Hash"`
	Data                  *VoteData `json:"Data"`
}

type VoteData struct {
	Hash            string `json:"Hash"`
	Slot            int    `json:"Slot"`
	FromEpoch       int    `json:"FromEpoch"`
	FromHash        string `json:"FromHash"`
	ToEpoch         int    `json:"ToEpoch"`
	ToHash          string `json:"ToHash"`
	BeaconBlockHash string `json:"BeaconBlockHash"`
	Nonce           string `json:"Nonce"`
}
