package primitives

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-bitfield"
)

// SerializableState is a serializable copy of the state
type SerializableState struct {
	// CoinsState keeps if accounts balances and transactions.
	CoinsState *CoinsStateSerializable
	// ValidatorRegistry keeps track of validators in the state.
	ValidatorRegistry []*Validator `ssz-max:"1099511627776"`

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
	ProposerQueue []uint64 `ssz-max:"1099511627776"`

	PreviousEpochVoteAssignments []uint64 `ssz-max:"1099511627776"`
	CurrentEpochVoteAssignments  []uint64 `ssz-max:"1099511627776"`

	// NextProposerQueue is the queue of validators scheduled to create a block
	// in the next epoch.
	NextProposerQueue []uint64 `ssz-max:"1099511627776"`

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
	CurrentEpochVotes []*AcceptedVoteInfo `ssz-max:"1099511627776"`

	// PreviousJustifiedEpoch is the second-to-last epoch that >2/3 of validators
	// voted for.
	PreviousJustifiedEpoch uint64

	// PreviousJustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	PreviousJustifiedEpochHash [32]byte `ssz-size:"32"`

	// PreviousEpochVotes are votes where the FromEpoch matches PreviousJustifiedEpoch.
	PreviousEpochVotes []*AcceptedVoteInfo `ssz-max:"1099511627776"`

	// CurrentManagers are current managers of the governance funds.
	CurrentManagers [][20]byte `ssz-max:"5"`

	// ManagerReplacement is a bitfield where the bits of the managers to replace are 1.
	ManagerReplacement bitfield.Bitlist `ssz:"bitlist" ssz-max:"2048"`

	// Governance represents current votes state
	Governance *GovernanceSerializable

	VoteEpoch          uint64
	VoteEpochStartSlot uint64
	VotingState        uint64

	LastPaidSlot uint64
}

// StateValidatorsInfo returns the state validators information.
type StateValidatorsInfo struct {
	Validators  []*Validator
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

// State is the state of consensus in the blockchain.
type State struct {
	// CoinsState keeps if accounts balances and transactions.
	CoinsState CoinsState
	// ValidatorRegistry keeps track of validators in the state.
	ValidatorRegistry []*Validator

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
	CurrentEpochVotes []*AcceptedVoteInfo

	// PreviousJustifiedEpoch is the second-to-last epoch that >2/3 of validators
	// voted for.
	PreviousJustifiedEpoch uint64

	// PreviousJustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	PreviousJustifiedEpochHash chainhash.Hash

	// PreviousEpochVotes are votes where the FromEpoch matches PreviousJustifiedEpoch.
	PreviousEpochVotes []*AcceptedVoteInfo

	// CurrentManagers are current managers of the governance funds.
	CurrentManagers [][20]byte

	// ManagerReplacement is a bitfield where the bits of the managers to replace are 1.
	ManagerReplacement bitfield.Bitlist

	// Governance represents current votes state
	Governance Governance

	VoteEpoch          uint64
	VoteEpochStartSlot uint64
	VotingState        uint64

	LastPaidSlot uint64
}

// ToSerializable converts the struct to a serializable struct
func (s *State) ToSerializable() SerializableState {
	serCoin := s.CoinsState.ToSerializable()
	serGov := s.Governance.ToSerializable()
	ser := SerializableState{
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
func (s *State) FromSerializable(ser *SerializableState) {
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
func (s *State) Marshal() ([]byte, error) {
	ser := s.ToSerializable()
	b, err := ser.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Unmarshal decodes the data.
func (s *State) Unmarshal(b []byte) error {
	ser := new(SerializableState)
	err := ser.UnmarshalSSZ(b)
	if err != nil {
		return err
	}
	s.FromSerializable(ser)
	return nil

}

// GetValidatorIndicesActiveAt gets validator indices where the validator is active at a certain slot.
func (s *State) GetValidatorIndicesActiveAt(epoch uint64) []uint64 {
	vals := make([]uint64, 0, len(s.ValidatorRegistry))
	for i, v := range s.ValidatorRegistry {
		if v.IsActiveAtEpoch(epoch) {
			vals = append(vals, uint64(i))
		}
	}

	return vals
}

// GetValidators returns the validator information at current state
func (s *State) GetValidators() StateValidatorsInfo {
	validators := StateValidatorsInfo{
		Validators:  s.ValidatorRegistry,
		Active:      int64(0),
		PendingExit: int64(0),
		PenaltyExit: int64(0),
		Exited:      int64(0),
		Starting:    int64(0),
	}
	for _, v := range s.ValidatorRegistry {
		switch v.Status {
		case StatusActive:
			validators.Active++
		case StatusActivePendingExit:
			validators.PendingExit++
		case StatusExitedWithPenalty:
			validators.PenaltyExit++
		case StatusExitedWithoutPenalty:
			validators.Exited++
		case StatusStarting:
			validators.Starting++
		}
	}
	return validators
}

// GetValidatorsForAccount returns the validator information at current state from a defined account
func (s *State) GetValidatorsForAccount(acc []byte) StateValidatorsInfo {
	var account [20]byte
	copy(account[:], acc)
	validators := StateValidatorsInfo{
		Validators:  []*Validator{},
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
			case StatusActive:
				validators.Active++
			case StatusActivePendingExit:
				validators.PendingExit++
			case StatusExitedWithPenalty:
				validators.PenaltyExit++
			case StatusExitedWithoutPenalty:
				validators.Exited++
			case StatusStarting:
				validators.Starting++
			}
		}
	}
	return validators
}

// Copy returns a copy of the state.
func (s *State) Copy() State {
	s2 := *s

	s2.CoinsState = s.CoinsState.Copy()

	s2.ValidatorRegistry = make([]*Validator, len(s.ValidatorRegistry))

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

	s2.CurrentEpochVotes = make([]*AcceptedVoteInfo, len(s.CurrentEpochVotes))
	for i, c := range s.CurrentEpochVotes {
		cv := c.Copy()
		s2.CurrentEpochVotes[i] = &cv
	}

	s2.PreviousEpochVotes = make([]*AcceptedVoteInfo, len(s.PreviousEpochVotes))
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
	return s2
}
