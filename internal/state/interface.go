package state

import (
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type State interface {
	ProcessSlot(previousBlockRoot chainhash.Hash)
	ProcessSlots(requestedSlot uint64, view BlockView) ([]*primitives.EpochReceipt, error)
	ProcessBlock(b *primitives.Block) error
	ProcessVote(v *primitives.MultiValidatorVote, proposerIndex uint64) error
	ProcessEpochTransition() ([]*primitives.EpochReceipt, error)

	CheckBlockSignature(b *primitives.Block) error
	IsProposerSlashingValid(ps *primitives.ProposerSlashing) (uint64, error)
	IsVoteSlashingValid(vs *primitives.VoteSlashing) ([]uint64, error)
	IsRANDAOSlashingValid(rs *primitives.RANDAOSlashing) (uint64, error)
	IsExitValid(exit *primitives.Exit) error
	IsDepositValid(deposit *primitives.Deposit) error
	IsVoteValid(v *primitives.MultiValidatorVote) error
	IsPartialExitValid(p *primitives.PartialExit) error

	AreDepositsValid(deposits []*primitives.Deposit) error

	ActivateValidator(index uint64) error
	InitiateValidatorExit(index uint64) error
	ExitValidator(index uint64, status uint64) error
	UpdateValidatorStatus(index uint64, status uint64) error

	ApplyRANDAOSlashing(rs *primitives.RANDAOSlashing) error
	ApplyTransactionSingle(tx *primitives.Tx, blockWithdrawalAddress [20]byte) error
	ApplyProposerSlashing(ps *primitives.ProposerSlashing) error
	ApplyVoteSlashing(vs *primitives.VoteSlashing) error
	ApplyExit(exit *primitives.Exit) error
	ApplyDeposit(deposit *primitives.Deposit) error
	ApplyPartialExit(p *primitives.PartialExit) error
	SetSlot(slot uint64)

	ApplyMultiTransactionSingle(txs []*primitives.Tx, blockWithdrawalAddress [20]byte) error
	ApplyMultiDeposit(deposits []*primitives.Deposit) error

	Copy() State
	ToSerializable() *primitives.SerializableState
	FromSerializable(ser *primitives.SerializableState)
	Marshal() ([]byte, error)
	Unmarshal(b []byte) error

	GetVoteCommittee(slot uint64) ([]uint64, error)
	GetProposerPublicKey(b *primitives.Block) (common.PublicKey, error)
	GetRecentBlockHash(slotToGet uint64) chainhash.Hash
	GetTotalBalances() uint64
	GetEffectiveBalance(index uint64) uint64
	GetValidatorIndicesActiveAt(epoch uint64) []uint64
	GetValidators() ValidatorsInfo
	GetValidatorsForAccount(acc []byte) ValidatorsInfo
	GetCoinsState() primitives.CoinsState
	GetValidatorRegistry() []*primitives.Validator
	GetProposerQueue() []uint64
	GetSlot() uint64
	GetEpochIndex() uint64
	GetFinalizedEpoch() uint64
	GetJustifiedEpoch() uint64
	GetJustifiedEpochHash() chainhash.Hash
}

func (s *state) GetCoinsState() primitives.CoinsState {
	return s.CoinsState
}

func (s *state) GetValidatorRegistry() []*primitives.Validator {
	return s.ValidatorRegistry
}

func (s *state) GetProposerQueue() []uint64 {
	return s.ProposerQueue
}

func (s *state) GetSlot() uint64 {
	return s.Slot
}

func (s *state) GetEpochIndex() uint64 {
	return s.EpochIndex
}

func (s *state) GetFinalizedEpoch() uint64 {
	return s.FinalizedEpoch
}

func (s *state) GetJustifiedEpoch() uint64 {
	return s.JustifiedEpoch
}

func (s *state) GetJustifiedEpochHash() chainhash.Hash {
	return s.JustifiedEpochHash
}

var _ State = &state{}
