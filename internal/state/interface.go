package state

import (
	"github.com/olympus-protocol/ogen/internal/logger"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type State interface {
	ProcessSlot(p *params.ChainParams, previousBlockRoot chainhash.Hash)
	ProcessSlots(requestedSlot uint64, view BlockView, p *params.ChainParams, log logger.Logger) ([]*primitives.EpochReceipt, error)
	GetEffectiveBalance(index uint64, p *params.ChainParams) uint64
	getActiveBalance(_ *params.ChainParams) uint64
	ActivateValidator(index uint64) error
	InitiateValidatorExit(index uint64) error
	ExitValidator(index uint64, status uint64, p *params.ChainParams) error
	UpdateValidatorStatus(index uint64, status uint64, p *params.ChainParams) error
	updateValidatorRegistry(p *params.ChainParams) error
	GetRecentBlockHash(slotToGet uint64, p *params.ChainParams) chainhash.Hash
	GetTotalBalances() uint64
	NextVoteEpoch(newState uint64)
	CheckForVoteTransitions(p *params.ChainParams)
	ProcessEpochTransition(p *params.ChainParams, _ logger.Logger) ([]*primitives.EpochReceipt, error)
	IsGovernanceVoteValid(vote *primitives.GovernanceVote, p *params.ChainParams) error
	ProcessGovernanceVote(vote *primitives.GovernanceVote, p *params.ChainParams) error
	ApplyTransactionSingle(tx *primitives.Tx, blockWithdrawalAddress [20]byte, p *params.ChainParams) error
	ApplyTransactionMulti(tx *primitives.TxMulti, blockWithdrawalAddress [20]byte, p *params.ChainParams) error
	IsProposerSlashingValid(ps *primitives.ProposerSlashing) (uint64, error)
	ApplyProposerSlashing(ps *primitives.ProposerSlashing, p *params.ChainParams) error
	IsVoteSlashingValid(vs *primitives.VoteSlashing, p *params.ChainParams) ([]uint64, error)
	ApplyVoteSlashing(vs *primitives.VoteSlashing, p *params.ChainParams) error
	IsRANDAOSlashingValid(rs *primitives.RANDAOSlashing) (uint64, error)
	ApplyRANDAOSlashing(rs *primitives.RANDAOSlashing, p *params.ChainParams) error
	GetVoteCommittee(slot uint64, p *params.ChainParams) ([]uint64, error)
	IsExitValid(exit *primitives.Exit) error
	ApplyExit(exit *primitives.Exit) error
	IsDepositValid(deposit *primitives.Deposit, params *params.ChainParams) error
	ApplyDeposit(deposit *primitives.Deposit, p *params.ChainParams) error
	IsVoteValid(v *primitives.MultiValidatorVote, p *params.ChainParams) error
	ProcessVote(v *primitives.MultiValidatorVote, p *params.ChainParams, proposerIndex uint64) error
	GetProposerPublicKey(b *primitives.Block, p *params.ChainParams) (bls_interface.PublicKey, error)
	CheckBlockSignature(b *primitives.Block, p *params.ChainParams) error
	ProcessBlock(b *primitives.Block, p *params.ChainParams) error
	ToSerializable() *SerializableState
	FromSerializable(ser *SerializableState)
	Marshal() ([]byte, error)
	Unmarshal(b []byte) error
	GetValidatorIndicesActiveAt(epoch uint64) []uint64
	GetValidators() ValidatorsInfo
	GetValidatorsForAccount(acc []byte) ValidatorsInfo
	Copy() State
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
