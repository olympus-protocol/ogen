package primitives

import (
	"bytes"
	"encoding/binary"
)

const (
	GovernanceStateActive uint8 = iota
	GovernanceStateVoting
)

// StateValidatorsInfo returns the state validators information.
type StateValidatorsInfo struct {
	Validators  []*Validator `ssz-max:"16777216"`
	Active      uint64
	PendingExit uint64
	PenaltyExit uint64
	Exited      uint64
	Starting    uint64
}

type AccountTxs struct {
	Amount uint64
	Txs    [][]byte `ssz-size:"?,32" ssz-max:"16777216"`
}

// State is the state of consensus in the blockchain.
type State struct {
	CoinsState *CoinsState

	// ValidatorRegistry keeps track of validators in the state.
	ValidatorRegistry []*Validator `ssz-max:"16777216"`

	// LatestValidatorRegistryChange keeps track of the last time the validator
	// registry was changed. We only want to update the registry if a block was
	// finalized since the last time it was changed, so we keep track of that
	// here.
	LatestValidatorRegistryChange uint64

	// RANDAO for figuring out the proposer queue. We don't want any one validator
	// to have influence over the RANDAO, so we have each proposer contribute.
	RANDAO []byte `ssz-size:"32"`

	// NextRANDAO is the RANDAO currently being created. Every time a block is
	// created, we XOR the 32 least-significant bytes of the RandaoReveal with this
	// value to update it.
	NextRANDAO []byte `ssz-size:"32"`

	// Slot is the last slot ProcessSlot was called for.
	Slot uint64

	// EpochIndex is the last epoch ProcessEpoch was called for.
	EpochIndex uint64

	// ProposerQueue is the queue of validators scheduled to create a block.
	ProposerQueue [][]byte `ssz-size:"?,4" ssz-max:"16777216"`

	PreviousEpochVoteAssignments [][]byte `ssz-size:"?,4" ssz-max:"16777216"`
	CurrentEpochVoteAssignments  [][]byte `ssz-size:"?,4" ssz-max:"16777216"`

	// NextProposerQueue is the queue of validators scheduled to create a block
	// in the next epoch.
	NextProposerQueue [][]byte `ssz-size:"?,4" ssz-max:"16777216"`

	// JustifiedBitfield is a bitfield where the nth least significant bit
	// represents whether the nth last epoch was justified.
	JustificationBitfield uint64

	// FinalizedEpoch is the epoch that was finalized.
	FinalizedEpoch uint64

	// LastBlockHashes is the last LastBlockHashesSize block hashes.
	LatestBlockHashes [][]byte `ssz-size:"?,32" ssz-max:"16777216"`

	// JustifiedEpoch is the last epoch that >2/3 of validators voted for.
	JustifiedEpoch uint64

	// JustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	JustifiedEpochHash []byte `ssz-size:"32"`

	// CurrentEpochVotes are votes that are being submitted where
	// the source epoch matches justified epoch.
	CurrentEpochVotes []*AcceptedVoteInfo `ssz-max:"16777216"`

	// PreviousJustifiedEpoch is the second-to-last epoch that >2/3 of validators
	// voted for.
	PreviousJustifiedEpoch uint64

	// PreviousJustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	PreviousJustifiedEpochHash []byte `ssz-size:"32"`

	// PreviousEpochVotes are votes where the FromEpoch matches PreviousJustifiedEpoch.
	PreviousEpochVotes []*AcceptedVoteInfo `ssz-max:"16777216"`

	// CurrentManagers are current managers of the governance funds.
	CurrentManagers [][]byte `ssz-size:"?,20" ssz-max:"16777216"`

	// ManagerReplacement is a bitfield where the bits of the managers to replace are 1.
	ManagerReplacement []byte `ssz-size:"32"`

	// ReplaceVotes are votes to start the community-override functionality. Each address
	// in here must have at least 100 POLIS and once that accounts for >=30% of the supply,
	// a community voting round starts.
	// For a voting period, the hash is set to the proposed community vote.
	// For a non-voting period, the hash is 0.
	//ReplaceVotes map[[20]byte]chainhash.Hash

	// CommunityVotes is set during a voting period to keep track of the
	// possible votes.
	//CommunityVotes map[chainhash.Hash]CommunityVoteData

	VoteEpoch          uint64
	VoteEpochStartSlot uint64
	VotingState        uint8

	LastPaidSlot uint64
}

// GetValidatorIndicesActiveAt gets validator indices where the validator is active at a certain slot.
func (s *State) GetValidatorIndicesActiveAt(epoch uint64) [][]byte {
	vals := make([][]byte, 0, len(s.ValidatorRegistry))
	for i, v := range s.ValidatorRegistry {
		if v.IsActiveAtEpoch(epoch) {
			val := []byte{}
			binary.LittleEndian.PutUint32(val, uint32(i))
			vals = append(vals, val)
		}
	}
	return vals
}

// GetValidators returns the validator information at current state
func (s *State) GetValidators() StateValidatorsInfo {
	validators := StateValidatorsInfo{
		Validators:  s.ValidatorRegistry,
		Active:      uint64(0),
		PendingExit: uint64(0),
		PenaltyExit: uint64(0),
		Exited:      uint64(0),
		Starting:    uint64(0),
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
	validators := StateValidatorsInfo{
		Validators:  []*Validator{},
		Active:      uint64(0),
		PendingExit: uint64(0),
		PenaltyExit: uint64(0),
		Exited:      uint64(0),
		Starting:    uint64(0),
	}
	for _, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PayeeAddress, acc) {
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
