package state

import (
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// SerializableState is a serializable copy of the state
type SerializableState struct {
	// CoinsState keeps if accounts balances and transactions.
	CoinsState *primitives.CoinsStateSerializable
	// ValidatorRegistry keeps track of validators in the state.
	ValidatorRegistry []*primitives.Validator `ssz-max:"2097152"`

	// LatestValidatorRegistryChange keeps track of the last time the validator
	// registry was changed. We only want to update the registry if a block was
	// finalized since the last time it was changed, so we keep track of that
	// here.
	LatestValidatorRegistryChange uint64

	// RANDAO for figuring out the proposer queue. We don't want any one validator
	// to have influence over the RANDAO, so we have each proposer contribute.
	RANDAO [32]byte `ssz-size:"32"`

	// NextRANDAO is the RANDAO currently being created. Every time a block is
	// created, we XOR the 32 least-significant bytes of the RandaoReveal with this
	// value to update it.
	NextRANDAO [32]byte `ssz-size:"32"`

	// Slot is the last slot ProcessSlot was called for.
	Slot uint64

	// EpochIndex is the last epoch ProcessEpoch was called for.
	EpochIndex uint64

	// ProposerQueue is the queue of validators scheduled to create a block.
	ProposerQueue []uint64 `ssz-max:"2097152"`

	PreviousEpochVoteAssignments []uint64 `ssz-max:"2097152"`
	CurrentEpochVoteAssignments  []uint64 `ssz-max:"2097152"`

	// NextProposerQueue is the queue of validators scheduled to create a block
	// in the next epoch.
	NextProposerQueue []uint64 `ssz-max:"2097152"`

	// JustifiedBitfield is a bitfield where the nth least significant bit
	// represents whether the nth last epoch was justified.
	JustificationBitfield uint64

	// FinalizedEpoch is the epoch that was finalized.
	FinalizedEpoch uint64

	// LastBlockHashes is the last LastBlockHashesSize block hashes.
	LatestBlockHashes [][32]byte `ssz-max:"64"`

	// JustifiedEpoch is the last epoch that >2/3 of validators voted for.
	JustifiedEpoch uint64

	// JustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	JustifiedEpochHash [32]byte `ssz-size:"32"`

	// CurrentEpochVotes are votes that are being submitted where
	// the source epoch matches justified epoch.
	CurrentEpochVotes []*primitives.AcceptedVoteInfo `ssz-max:"2097152"`

	// PreviousJustifiedEpoch is the second-to-last epoch that >2/3 of validators
	// voted for.
	PreviousJustifiedEpoch uint64

	// PreviousJustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	PreviousJustifiedEpochHash [32]byte `ssz-size:"32"`

	// PreviousEpochVotes are votes where the FromEpoch matches PreviousJustifiedEpoch.
	PreviousEpochVotes []*primitives.AcceptedVoteInfo `ssz-max:"2097152"`

	// CurrentManagers are current managers of the governance funds.
	CurrentManagers [][20]byte `ssz-max:"5"`

	// ManagerReplacement is a bitfield where the bits of the managers to replace are 1.
	ManagerReplacement bitfield.Bitlist `ssz:"bitlist" ssz-max:"2048"`

	// Governance represents current votes state
	Governance *primitives.GovernanceSerializable

	VoteEpoch          uint64
	VoteEpochStartSlot uint64
	VotingState        uint64

	LastPaidSlot uint64
}

// ValidatorsInfo returns the state validators information.
type ValidatorsInfo struct {
	Validators  []*primitives.Validator
	Active      int64
	PendingExit int64
	PenaltyExit int64
	Exited      int64
	Starting    int64
}

// LastBlockHashesSize is the size of the last block hashes.
const LastBlockHashesSize = 8

const (
	// GovernanceStateActive is the enum to an active Governance voting state
	GovernanceStateActive uint64 = iota
	// GovernanceStateVoting is the enum to an voting Governance voting state
	GovernanceStateVoting
)

// state is the state of consensus in the blockchain.
type state struct {
	// CoinsState keeps if accounts balances and transactions.
	CoinsState primitives.CoinsState
	// ValidatorRegistry keeps track of validators in the state.
	ValidatorRegistry []*primitives.Validator

	// LatestValidatorRegistryChange keeps track of the last time the validator
	// registry was changed. We only want to update the registry if a block was
	// finalized since the last time it was changed, so we keep track of that
	// here.
	LatestValidatorRegistryChange uint64

	// RANDAO for figuring out the proposer queue. We don't want any one validator
	// to have influence over the RANDAO, so we have each proposer contribute.
	RANDAO chainhash.Hash

	// NextRANDAO is the RANDAO currently being created. Every time a block is
	// created, we XOR the 32 least-significant bytes of the RandaoReveal with this
	// value to update it.
	NextRANDAO chainhash.Hash

	// Slot is the last slot ProcessSlot was called for.
	Slot uint64

	// EpochIndex is the last epoch ProcessEpoch was called for.
	EpochIndex uint64

	// ProposerQueue is the queue of validators scheduled to create a block.
	ProposerQueue []uint64

	PreviousEpochVoteAssignments []uint64
	CurrentEpochVoteAssignments  []uint64

	// NextProposerQueue is the queue of validators scheduled to create a block
	// in the next epoch.
	NextProposerQueue []uint64

	// JustifiedBitfield is a bitfield where the nth least significant bit
	// represents whether the nth last epoch was justified.
	JustificationBitfield uint64

	// FinalizedEpoch is the epoch that was finalized.
	FinalizedEpoch uint64

	// LastBlockHashes is the last LastBlockHashesSize block hashes.
	LatestBlockHashes [][32]byte

	// JustifiedEpoch is the last epoch that >2/3 of validators voted for.
	JustifiedEpoch uint64

	// JustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	JustifiedEpochHash chainhash.Hash

	// CurrentEpochVotes are votes that are being submitted where
	// the source epoch matches justified epoch.
	CurrentEpochVotes []*primitives.AcceptedVoteInfo

	// PreviousJustifiedEpoch is the second-to-last epoch that >2/3 of validators
	// voted for.
	PreviousJustifiedEpoch uint64

	// PreviousJustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	PreviousJustifiedEpochHash chainhash.Hash

	// PreviousEpochVotes are votes where the FromEpoch matches PreviousJustifiedEpoch.
	PreviousEpochVotes []*primitives.AcceptedVoteInfo

	// CurrentManagers are current managers of the governance funds.
	CurrentManagers [][20]byte

	// ManagerReplacement is a bitfield where the bits of the managers to replace are 1.
	ManagerReplacement bitfield.Bitlist

	// Governance represents current votes state
	Governance primitives.Governance

	VoteEpoch          uint64
	VoteEpochStartSlot uint64
	VotingState        uint64

	LastPaidSlot uint64
}

// ToSerializable converts the struct to a serializable struct
func (s *state) ToSerializable() *SerializableState {
	serCoin := s.CoinsState.ToSerializable()
	serGov := s.Governance.ToSerializable()
	ser := &SerializableState{
		CoinsState:                    &serCoin,
		ValidatorRegistry:             s.ValidatorRegistry,
		LatestValidatorRegistryChange: s.LatestValidatorRegistryChange,
		RANDAO:                        s.RANDAO,
		NextRANDAO:                    s.NextRANDAO,
		Slot:                          s.Slot,
		EpochIndex:                    s.EpochIndex,
		ProposerQueue:                 s.ProposerQueue,
		PreviousEpochVoteAssignments:  s.PreviousEpochVoteAssignments,
		CurrentEpochVoteAssignments:   s.CurrentEpochVoteAssignments,
		NextProposerQueue:             s.NextProposerQueue,
		JustificationBitfield:         s.JustificationBitfield,
		FinalizedEpoch:                s.FinalizedEpoch,
		LatestBlockHashes:             s.LatestBlockHashes,
		JustifiedEpoch:                s.JustifiedEpoch,
		JustifiedEpochHash:            s.JustifiedEpochHash,
		CurrentEpochVotes:             s.CurrentEpochVotes,
		PreviousJustifiedEpoch:        s.PreviousJustifiedEpoch,
		PreviousJustifiedEpochHash:    s.PreviousJustifiedEpochHash,
		PreviousEpochVotes:            s.PreviousEpochVotes,
		CurrentManagers:               s.CurrentManagers,
		ManagerReplacement:            s.ManagerReplacement,
		Governance:                    &serGov,
		VoteEpoch:                     s.VoteEpoch,
		VoteEpochStartSlot:            s.VoteEpochStartSlot,
		VotingState:                   s.VotingState,
		LastPaidSlot:                  s.LastPaidSlot,
	}
	return ser
}

// FromSerializable converts the struct to a serializable struct
func (s *state) FromSerializable(ser *SerializableState) {
	s.ValidatorRegistry = ser.ValidatorRegistry
	s.LatestValidatorRegistryChange = ser.LatestValidatorRegistryChange
	s.RANDAO = ser.RANDAO
	s.NextRANDAO = ser.NextRANDAO
	s.Slot = ser.Slot
	s.EpochIndex = ser.EpochIndex
	s.ProposerQueue = ser.ProposerQueue
	s.PreviousEpochVoteAssignments = ser.PreviousEpochVoteAssignments
	s.CurrentEpochVoteAssignments = ser.CurrentEpochVoteAssignments
	s.NextProposerQueue = ser.NextProposerQueue
	s.JustificationBitfield = ser.JustificationBitfield
	s.FinalizedEpoch = ser.FinalizedEpoch
	s.LatestBlockHashes = ser.LatestBlockHashes
	s.JustifiedEpoch = ser.JustifiedEpoch
	s.JustifiedEpochHash = ser.JustifiedEpochHash
	s.CurrentEpochVotes = ser.CurrentEpochVotes
	s.PreviousJustifiedEpoch = ser.PreviousJustifiedEpoch
	s.PreviousJustifiedEpochHash = ser.PreviousJustifiedEpochHash
	s.PreviousEpochVotes = ser.PreviousEpochVotes
	s.CurrentManagers = ser.CurrentManagers
	s.ManagerReplacement = ser.ManagerReplacement
	s.VoteEpoch = ser.VoteEpoch
	s.VoteEpochStartSlot = ser.VoteEpochStartSlot
	s.VotingState = ser.VotingState
	s.LastPaidSlot = ser.LastPaidSlot
	s.CoinsState.FromSerializable(ser.CoinsState)
	s.Governance.FromSerializable(ser.Governance)
	return
}

// Marshal encodes the data.
func (s *state) Marshal() ([]byte, error) {
	ser := s.ToSerializable()
	b, err := ser.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Unmarshal decodes the data.
func (s *state) Unmarshal(b []byte) error {
	ser := new(SerializableState)
	err := ser.UnmarshalSSZ(b)
	if err != nil {
		return err
	}
	s.FromSerializable(ser)
	return nil

}

// GetValidatorIndicesActiveAt gets validator indices where the validator is active at a certain slot.
func (s *state) GetValidatorIndicesActiveAt(epoch uint64) []uint64 {
	vals := make([]uint64, 0, len(s.ValidatorRegistry))
	for i, v := range s.ValidatorRegistry {
		if v.IsActiveAtEpoch(epoch) {
			vals = append(vals, uint64(i))
		}
	}

	return vals
}

// GetValidators returns the validator information at current state
func (s *state) GetValidators() ValidatorsInfo {
	validators := ValidatorsInfo{
		Validators:  s.ValidatorRegistry,
		Active:      int64(0),
		PendingExit: int64(0),
		PenaltyExit: int64(0),
		Exited:      int64(0),
		Starting:    int64(0),
	}
	for _, v := range s.ValidatorRegistry {
		switch v.Status {
		case primitives.StatusActive:
			validators.Active++
		case primitives.StatusActivePendingExit:
			validators.PendingExit++
		case primitives.StatusExitedWithPenalty:
			validators.PenaltyExit++
		case primitives.StatusExitedWithoutPenalty:
			validators.Exited++
		case primitives.StatusStarting:
			validators.Starting++
		}
	}
	return validators
}

// GetValidatorsForAccount returns the validator information at current state from a defined account
func (s *state) GetValidatorsForAccount(acc []byte) ValidatorsInfo {
	var account [20]byte
	copy(account[:], acc)
	validators := ValidatorsInfo{
		Validators:  []*primitives.Validator{},
		Active:      int64(0),
		PendingExit: int64(0),
		PenaltyExit: int64(0),
		Exited:      int64(0),
		Starting:    int64(0),
	}
	for _, v := range s.ValidatorRegistry {
		if v.PayeeAddress == account {
			validators.Validators = append(validators.Validators, v)
			switch v.Status {
			case primitives.StatusActive:
				validators.Active++
			case primitives.StatusActivePendingExit:
				validators.PendingExit++
			case primitives.StatusExitedWithPenalty:
				validators.PenaltyExit++
			case primitives.StatusExitedWithoutPenalty:
				validators.Exited++
			case primitives.StatusStarting:
				validators.Starting++
			}
		}
	}
	return validators
}

// Copy returns a copy of the state.
func (s *state) Copy() State {
	s2 := *s

	s2.CoinsState = s.CoinsState.Copy()

	s2.ValidatorRegistry = make([]*primitives.Validator, len(s.ValidatorRegistry))

	for i, c := range s.ValidatorRegistry {
		v := c.Copy()
		s2.ValidatorRegistry[i] = &v
	}

	s2.ProposerQueue = make([]uint64, len(s.ProposerQueue))
	for i, c := range s.ProposerQueue {
		s2.ProposerQueue[i] = c
	}

	s2.NextProposerQueue = make([]uint64, len(s.NextProposerQueue))
	for i, c := range s.NextProposerQueue {
		s2.NextProposerQueue[i] = c
	}

	s2.CurrentEpochVoteAssignments = make([]uint64, len(s.CurrentEpochVoteAssignments))
	for i, c := range s.CurrentEpochVoteAssignments {
		s2.CurrentEpochVoteAssignments[i] = c
	}

	s2.PreviousEpochVoteAssignments = make([]uint64, len(s.PreviousEpochVoteAssignments))
	for i, c := range s.PreviousEpochVoteAssignments {
		s2.PreviousEpochVoteAssignments[i] = c
	}

	s2.LatestBlockHashes = make([][32]byte, len(s.LatestBlockHashes))
	for i, c := range s.LatestBlockHashes {
		s2.LatestBlockHashes[i] = c
	}

	s2.CurrentEpochVotes = make([]*primitives.AcceptedVoteInfo, len(s.CurrentEpochVotes))
	for i, c := range s.CurrentEpochVotes {
		cv := c.Copy()
		s2.CurrentEpochVotes[i] = &cv
	}

	s2.PreviousEpochVotes = make([]*primitives.AcceptedVoteInfo, len(s.PreviousEpochVotes))
	for i, c := range s.PreviousEpochVotes {
		cv := c.Copy()
		s2.PreviousEpochVotes[i] = &cv
	}

	s2.CurrentManagers = make([][20]byte, len(s.CurrentManagers))
	copy(s2.CurrentManagers, s.CurrentManagers)

	s2.ManagerReplacement = bitfield.NewBitlist(s.ManagerReplacement.Len())
	for i, b := range s.ManagerReplacement {
		s2.ManagerReplacement[i] = b
	}

	s2.Governance = s.Governance.Copy()
	return &s2
}

func NewState(cs primitives.CoinsState, validators []*primitives.Validator, genHash chainhash.Hash, p *params.ChainParams) State {
	s := &state{
		CoinsState:                    cs,
		ValidatorRegistry:             validators,
		LatestValidatorRegistryChange: 0,
		RANDAO:                        chainhash.Hash{},
		NextRANDAO:                    chainhash.Hash{},
		Slot:                      0,
		EpochIndex:                    0,
		JustificationBitfield:         0,
		JustifiedEpoch:                0,
		FinalizedEpoch:                0,
		LatestBlockHashes:             make([][32]byte, p.LatestBlockRootsLength),
		JustifiedEpochHash:            genHash,
		CurrentEpochVotes:             make([]*primitives.AcceptedVoteInfo, 0),
		PreviousJustifiedEpoch:        0,
		PreviousJustifiedEpochHash:    genHash,
		PreviousEpochVotes:            make([]*primitives.AcceptedVoteInfo, 0),
		CurrentManagers:               p.InitialManagers,
		VoteEpoch:                     0,
		VoteEpochStartSlot:            0,
		VotingState:                   GovernanceStateActive,
		LastPaidSlot:                  0,
	}
	activeValidators := s.GetValidatorIndicesActiveAt(0)
	s.ProposerQueue = DetermineNextProposers(chainhash.Hash{}, activeValidators, p)
	s.NextProposerQueue = DetermineNextProposers(chainhash.Hash{}, activeValidators, p)
	s.CurrentEpochVoteAssignments = Shuffle(chainhash.Hash{}, activeValidators)
	s.PreviousEpochVoteAssignments = Shuffle(chainhash.Hash{}, activeValidators)
	s.ManagerReplacement = bitfield.NewBitlist(uint64(len(s.CurrentManagers)) * 8)
	return s
}
