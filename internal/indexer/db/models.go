package db

import (
	"encoding/hex"
	"github.com/olympus-protocol/ogen/internal/indexer/graph/model"
	"time"
)

type Block struct {
	Hash      []byte `gorm:"primaryKey"`
	Height    uint64
	Slot      uint64
	Txs       []Tx
	Deposits  []Deposit
	Votes     []Vote
	Exits     []Exit
	RawBlock  []byte
	Timestamp uint64
}

func (b *Block) ToGQL() *model.Block {
	return &model.Block{
		Hash:      hex.EncodeToString(b.Hash),
		Height:    int(b.Height),
		Slot:      int(b.Slot),
		Timestamp: int(b.Timestamp),
		Txs:       b.TxsGQL(),
		Deposits:  b.DepositsGQL(),
		Votes:     b.VotesGQL(),
		Exits:     b.ExitsGQL(),
		RawBlock:  hex.EncodeToString(b.RawBlock),
	}
}

func (b *Block) DepositsGQL() []*model.Deposit {
	deposits := make([]*model.Deposit, len(b.Deposits))
	for i := range deposits {
		deposits[i] = b.Deposits[i].ToGQL()
	}
	return deposits
}

func (b *Block) TxsGQL() []*model.Tx {
	txs := make([]*model.Tx, len(b.Txs))
	for i := range txs {
		txs[i] = b.Txs[i].ToGQL()
	}
	return txs
}

func (b *Block) VotesGQL() []*model.Vote {
	votes := make([]*model.Vote, len(b.Votes))
	for i := range votes {
		votes[i] = b.Votes[i].ToGQL()
	}
	return votes
}

func (b *Block) ExitsGQL() []*model.Exit {
	exits := make([]*model.Exit, len(b.Exits))
	for i := range exits {
		exits[i] = b.Exits[i].ToGQL()
	}
	return exits
}

type BlockHeader struct {
	Hash                        []byte `gorm:"primaryKey"`
	Version                     uint64
	Nonce                       []byte
	Timestamp                   time.Time
	Slot                        uint64
	FeeAddress                  []byte
	PreviousBlockHash           []byte
	VotesMerkleRoot             []byte
	DeposistMerkleRoot          []byte
	ExitsMerkleRoot             []byte
	PartialExitsMerkleRoot      []byte
	CoinProofsMerkleRoot        []byte
	ExecutionsMerkleRoot        []byte
	TxsMerkleRoot               []byte
	VoteSlashingMerkleRoot      []byte
	RandaoSlashingMerkleRoot    []byte
	ProposerSlashingMerkleRoot  []byte
	GovernanceVotesMerkleRoot   []byte
	MultiSignatureTxsMerkleRoot []byte
}

func (b *BlockHeader) ToGQL() *model.BlockHeader {
	return &model.BlockHeader{
		Hash:                        hex.EncodeToString(b.Hash),
		Version:                     int(b.Version),
		Nonce:                       hex.EncodeToString(b.Nonce),
		Timestamp:                   b.Timestamp.String(),
		Slot:                        int(b.Slot),
		FeeAddress:                  hex.EncodeToString(b.FeeAddress),
		PreviousBlockHash:           hex.EncodeToString(b.PreviousBlockHash),
		VotesMerkleRoot:             hex.EncodeToString(b.VotesMerkleRoot),
		DepositsMerkleRoot:          hex.EncodeToString(b.DeposistMerkleRoot),
		ExitsMerkleRoot:             hex.EncodeToString(b.ExitsMerkleRoot),
		PartialExitsMerkleRoot:      hex.EncodeToString(b.PartialExitsMerkleRoot),
		CoinProofsMerkleRoot:        hex.EncodeToString(b.CoinProofsMerkleRoot),
		ExecutionsMerkleRoot:        hex.EncodeToString(b.ExecutionsMerkleRoot),
		TxsMerkleRoot:               hex.EncodeToString(b.TxsMerkleRoot),
		VoteSlashingMerkleRoot:      hex.EncodeToString(b.VoteSlashingMerkleRoot),
		RandaoSlashingMerkleRoot:    hex.EncodeToString(b.RandaoSlashingMerkleRoot),
		ProposerSlashingMerkleRoot:  hex.EncodeToString(b.ProposerSlashingMerkleRoot),
		GovernanceVotesMerkleRoot:   hex.EncodeToString(b.GovernanceVotesMerkleRoot),
		MultiSignatureTxsMerkleRoot: hex.EncodeToString(b.MultiSignatureTxsMerkleRoot),
	}
}

type Account struct {
	Account string `gorm:"primaryKey"`
	Balance uint64
	Nonce   uint64
}

func (a *Account) ToGQL() *model.Account {
	return &model.Account{
		Account: a.Account,
		Balance: int(a.Balance),
		Nonce:   int(a.Nonce),
	}
}

type CoinProofs struct {
	Hash          []byte `gorm:"primarykey"`
	Transaction   []byte
	RedeemAccount string
}

func (c *CoinProofs) ToGQL() *model.CoinProofs {
	return &model.CoinProofs{
		Hash:          hex.EncodeToString(c.Hash),
		Transaction:   hex.EncodeToString(c.Transaction),
		RedeemAccount: c.RedeemAccount,
	}
}

type Deposit struct {
	Hash      []byte `gorm:"primaryKey"`
	BlockHash []byte
	PublicKey []byte
	Data      DepositData `gorm:"foreignKey:Hash"`
}

func (d *Deposit) ToGQL() *model.Deposit {
	return &model.Deposit{
		Hash:      hex.EncodeToString(d.Hash),
		BlockHash: hex.EncodeToString(d.BlockHash),
		PublicKey: hex.EncodeToString(d.PublicKey),
		Data:      d.Data.ToGQL(),
	}
}

type DepositData struct {
	Hash              []byte
	PublicKey         []byte `gorm:"primaryKey"`
	ProofOfPossession []byte
	WithdrawalAddress []byte
}

func (d *DepositData) ToGQL() *model.DepositData {
	return &model.DepositData{
		Hash:              hex.EncodeToString(d.Hash),
		PublicKey:         hex.EncodeToString(d.PublicKey),
		ProofOfPossession: hex.EncodeToString(d.ProofOfPossession),
		WithdrawalAddress: hex.EncodeToString(d.WithdrawalAddress),
	}
}

type Epoch struct {
	Epoch         uint64 `gorm:"primaryKey"`
	Slot1         uint64
	Slot2         uint64
	Slot3         uint64
	Slot4         uint64
	Slot5         uint64
	ExpectedVotes uint64
	Finalized     bool
	Justified     bool
	Randao        []byte
}

func (e *Epoch) ToGQL() *model.Epoch {
	return &model.Epoch{
		Epoch:         int(e.Epoch),
		Slot1:         int(e.Slot1),
		Slot2:         int(e.Slot2),
		Slot3:         int(e.Slot3),
		Slot4:         int(e.Slot4),
		Slot5:         int(e.Slot5),
		ExpectedVotes: int(e.ExpectedVotes),
		Finalized:     e.Finalized,
		Justified:     e.Justified,
		Randao:        hex.EncodeToString(e.Randao),
	}
}

type Exit struct {
	Hash                []byte `gorm:"primarykey"`
	BlockHash           []byte
	ValidatorPublicKey  []byte
	WithdrawalPublicKey []byte
}

func (e *Exit) ToGQL() *model.Exit {
	return &model.Exit{
		Hash:                hex.EncodeToString(e.Hash),
		BlockHash:           hex.EncodeToString(e.BlockHash),
		ValidatorPublicKey:  hex.EncodeToString(e.ValidatorPublicKey),
		WithdrawalPublicKey: hex.EncodeToString(e.WithdrawalPublicKey),
	}
}

type PartialExit struct {
	Hash                []byte
	BlockHash           []byte
	ValidatorPublicKey  []byte
	WithdrawalPublicKey []byte
	Amount              uint64
}

type Slot struct {
	Epoch         uint64
	Slot          uint64 `gorm:"primaryKey"`
	BlockHash     []byte
	ProposerIndex uint64
	Proposed      bool
	VotesIncluded uint64
}

func (s *Slot) ToGQL() *model.Slot {
	return &model.Slot{
		Slot:          int(s.Slot),
		Epoch:         int(s.Epoch),
		BlockHash:     hex.EncodeToString(s.BlockHash),
		ProposerIndex: int(s.ProposerIndex),
		Proposed:      s.Proposed,
	}
}

type Tx struct {
	BlockHash         []byte
	Hash              []byte `gorm:"primarykey"`
	ToAddress         string
	FromPublicKeyHash string
	FromPublicKey     []byte
	Amount            uint64
	Nonce             uint64
	Fee               uint64
	Timestamp         uint64
}

func (t *Tx) ToGQL() *model.Tx {
	return &model.Tx{
		BlockHash:         hex.EncodeToString(t.BlockHash),
		Hash:              hex.EncodeToString(t.Hash),
		ToAddress:         t.ToAddress,
		FromPublicKeyHash: t.FromPublicKeyHash,
		FromPublicKey:     hex.EncodeToString(t.FromPublicKey),
		Amount:            int(t.Amount),
		Nonce:             int(t.Nonce),
		Fee:               int(t.Fee),
		Timestamp:         int(t.Timestamp),
	}
}

type Validator struct {
	Balance          uint64
	PubKey           []byte `gorm:"primaryKey"`
	PayeeAddress     string
	Status           uint64
	FirstActiveEpoch uint64
	LastActiveEpoch  uint64
}

func (v *Validator) ToGQL() *model.Validator {
	return &model.Validator{
		Balance:          int(v.Balance),
		Pubkey:           hex.EncodeToString(v.PubKey),
		PayeeAddress:     v.PayeeAddress,
		Status:           int(v.Status),
		FirstActiveEpoch: int(v.FirstActiveEpoch),
		LastActiveEpoch:  int(v.LastActiveEpoch),
	}
}

type Vote struct {
	BlockHash             []byte
	ParticipationBitfield []byte
	Hash                  []byte   `gorm:"primaryKey"`
	Data                  VoteData `gorm:"foreignKey:Hash"`
}

func (v *Vote) ToGQL() *model.Vote {
	return &model.Vote{
		BlockHash:             hex.EncodeToString(v.BlockHash),
		ParticipationBitfield: hex.EncodeToString(v.ParticipationBitfield),
		Hash:                  hex.EncodeToString(v.Hash),
		Data:                  v.Data.ToGQL(),
	}
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

func (v *VoteData) ToGQL() *model.VoteData {
	return &model.VoteData{
		Hash:            hex.EncodeToString(v.Hash),
		Slot:            int(v.Slot),
		FromEpoch:       int(v.FromEpoch),
		FromHash:        hex.EncodeToString(v.FromHash),
		ToEpoch:         int(v.ToEpoch),
		ToHash:          hex.EncodeToString(v.ToHash),
		BeaconBlockHash: hex.EncodeToString(v.BeaconBlockHash),
		Nonce:           hex.EncodeToString(v.Nonce),
	}
}

type State struct {
	Key             string `gorm:"primaryKey"`
	Raw             []byte
	LastBlock       []byte
	LastBlockHeight uint64
}

type AccountBalanceNotify struct {
	account string
	db      *Database
	notify  chan *model.Account
}

func (a *AccountBalanceNotify) Notify() {
	var initAccData Account
	res := a.db.DB.Where(&Account{Account: a.account}).First(&initAccData)
	if res.Error != nil {
		return
	}

	a.notify <- initAccData.ToGQL()
}

func NewAccountBalanceNotify(account string, channel chan *model.Account, db *Database) *AccountBalanceNotify {
	return &AccountBalanceNotify{db: db, notify: channel, account: account}
}

type TipNotify struct {
	db     *Database
	notify chan *model.Tip
}

func (t *TipNotify) Notify() {
	var data []*Validator

	t.db.DB.Find(&data)

	validators := make([]*model.Validator, len(data))

	for i := range validators {
		validators[i] = data[i].ToGQL()
	}

	var maxSlot int
	err := t.db.DB.Raw("select max(slot) from slots").Row().Scan(&maxSlot)
	if err != nil {
		return
	}

	var slot Slot
	res := t.db.DB.Where(&Slot{Slot: uint64(maxSlot)}).First(&slot)

	if res.Error != nil {
		return
	}

	var maxEpoch int
	err = t.db.DB.Raw("select max(epoch) from epoches").Row().Scan(&maxEpoch)
	if err != nil {
		return
	}

	var epoch Epoch
	res = t.db.DB.Where(&Epoch{Epoch: uint64(maxEpoch)}).First(&epoch)

	if res.Error != nil {
		return
	}

	var maxBlockHeight int
	err = t.db.DB.Raw("select max(height) from blocks").Row().Scan(&maxBlockHeight)
	if err != nil {
		return
	}

	var block Block
	res = t.db.DB.Where(&Block{Height: uint64(maxBlockHeight)}).First(&block)
	if res.Error != nil {
		return
	}

	epochGQL := epoch.ToGQL()

	participation, err := t.db.GetEpochParticipation(epochGQL.Epoch)
	if err != nil {
		return
	}

	epochGQL.Participation = participation

	t.notify <- &model.Tip{
		Slot:       slot.ToGQL(),
		Epoch:      epochGQL,
		Block:      block.ToGQL(),
		Validators: validators,
	}

	return
}

func NewTipNotify(channel chan *model.Tip, db *Database) *TipNotify {
	return &TipNotify{db: db, notify: channel}
}
