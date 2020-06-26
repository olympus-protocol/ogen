package primitives

import (
	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// CombineHashes combines two branches of the tree.
func CombineHashes(left *chainhash.Hash, right *chainhash.Hash) chainhash.Hash {
	return chainhash.HashH(append(left[:], right[:]...))
}

// EmptyTrees are empty trees for each level.
var EmptyTrees [256]chainhash.Hash

// EmptyTree is the hash of an empty tree.
var EmptyTree = chainhash.Hash{}

// UpdateWitness allows an executor to securely update the tree root so that only a single key is changed.
type UpdateWitness struct {
	Key             chainhash.Hash
	OldValue        chainhash.Hash
	NewValue        chainhash.Hash
	WitnessBitfield chainhash.Hash
	LastLevel       uint8
	Witnesses       []chainhash.Hash
}

// Copy returns a copy of the update witness.
func (uw *UpdateWitness) Copy() UpdateWitness {
	newUw := *uw

	newUw.Witnesses = make([]chainhash.Hash, len(uw.Witnesses))
	for i := range newUw.Witnesses {
		copy(newUw.Witnesses[i][:], uw.Witnesses[i][:])
	}

	return newUw
}

// VerificationWitness allows an executor to verify a specific node in the tree.
type VerificationWitness struct {
	Key             chainhash.Hash
	Value           chainhash.Hash
	WitnessBitfield chainhash.Hash
	Witnesses       []chainhash.Hash
	LastLevel       uint8
}

// Copy returns a copy of the update witness.
func (vw *VerificationWitness) Copy() VerificationWitness {
	newVw := *vw

	newVw.Witnesses = make([]chainhash.Hash, len(vw.Witnesses))
	for i := range newVw.Witnesses {
		copy(newVw.Witnesses[i][:], vw.Witnesses[i][:])
	}

	return newVw
}

// TransactionPackage is a way to update the state root without having the entire state.
type TransactionPackage struct {
	StartRoot     chainhash.Hash
	EndRoot       chainhash.Hash
	Updates       []UpdateWitness
	Verifications []VerificationWitness
}

// Copy copies the transaction package.
func (tp *TransactionPackage) Copy() TransactionPackage {
	newTp := *tp

	newTp.Updates = make([]UpdateWitness, len(tp.Updates))
	newTp.Verifications = make([]VerificationWitness, len(tp.Verifications))

	for i := range newTp.Updates {
		newTp.Updates[i] = tp.Updates[i].Copy()
	}

	for i := range newTp.Verifications {
		newTp.Verifications[i] = tp.Verifications[i].Copy()
	}

	return newTp
}

// StateValidatorsInfo returns the state validators information.
type StateValidatorsInfo struct {
	Validators  []Validator
	Active      int64
	PendingExit int64
	PenaltyExit int64
	Exited      int64
	Starting    int64
}

// LastBlockHashesSize is the size of the last block hashes.
const LastBlockHashesSize = 8

type GovernanceState uint8

const (
	GovernanceStateActive GovernanceState = iota
	GovernanceStateVoting
)

// State is the state of consensus in the blockchain.
type State struct {
	CoinsState CoinsState

	// ValidatorRegistry keeps track of validators in the state.
	ValidatorRegistry []Validator

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
	ProposerQueue []uint32

	PreviousEpochVoteAssignments []uint32
	CurrentEpochVoteAssignments  []uint32

	// NextProposerQueue is the queue of validators scheduled to create a block
	// in the next epoch.
	NextProposerQueue []uint32

	// JustifiedBitfield is a bitfield where the nth least significant bit
	// represents whether the nth last epoch was justified.
	JustificationBitfield uint64

	// FinalizedEpoch is the epoch that was finalized.
	FinalizedEpoch uint64

	// LastBlockHashes is the last LastBlockHashesSize block hashes.
	LatestBlockHashes []chainhash.Hash

	// JustifiedEpoch is the last epoch that >2/3 of validators voted for.
	JustifiedEpoch uint64

	// JustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	JustifiedEpochHash chainhash.Hash

	// CurrentEpochVotes are votes that are being submitted where
	// the source epoch matches justified epoch.
	CurrentEpochVotes []AcceptedVoteInfo

	// PreviousJustifiedEpoch is the second-to-last epoch that >2/3 of validators
	// voted for.
	PreviousJustifiedEpoch uint64

	// PreviousJustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	PreviousJustifiedEpochHash chainhash.Hash

	// PreviousEpochVotes are votes where the FromEpoch matches PreviousJustifiedEpoch.
	PreviousEpochVotes []AcceptedVoteInfo

	// CurrentManagers are current managers of the governance funds.
	CurrentManagers [][20]byte

	// ManagerReplacement is a bitfield where the bits of the managers to replace are 1.
	ManagerReplacement bitfield.Bitfield

	// ReplaceVotes are votes to start the community-override functionality. Each address
	// in here must have at least 100 POLIS and once that accounts for >=30% of the supply,
	// a community voting round starts.
	// For a voting period, the hash is set to the proposed community vote.
	// For a non-voting period, the hash is 0.
	ReplaceVotes map[[20]byte]chainhash.Hash

	// CommunityVotes is set during a voting period to keep track of the
	// possible votes.
	CommunityVotes map[chainhash.Hash]CommunityVoteData

	VoteEpoch          uint64
	VoteEpochStartSlot uint64
	VotingState        GovernanceState

	LastPaidSlot uint64
}

// Marshal encodes the data.
func (s *State) Marshal() ([]byte, error) {
	return nil, nil
}

// Unmarshal decodes the data.
func (s *State) Unmarshal(b []byte) error {
	return nil
}

// GetValidatorIndicesActiveAt gets validator indices where the validator is active at a certain slot.
func (s *State) GetValidatorIndicesActiveAt(epoch uint64) []uint32 {
	vals := make([]uint32, 0, len(s.ValidatorRegistry))
	for i, v := range s.ValidatorRegistry {
		if v.IsActiveAtEpoch(epoch) {
			vals = append(vals, uint32(i))
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
		Validators:  []Validator{},
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

// Hash calculates the hash of the state.
func (s *State) Hash() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(s)
	return chainhash.Hash(hash)
}

// Copy returns a copy of the state.
func (s *State) Copy() State {
	s2 := *s

	s2.CoinsState = s.CoinsState.Copy()
	s2.ValidatorRegistry = make([]Validator, len(s.ValidatorRegistry))

	for i, c := range s.ValidatorRegistry {
		s2.ValidatorRegistry[i] = c.Copy()
	}

	s2.ProposerQueue = make([]uint32, len(s.ProposerQueue))
	for i, c := range s.ProposerQueue {
		s2.ProposerQueue[i] = c
	}

	s2.NextProposerQueue = make([]uint32, len(s.NextProposerQueue))
	for i, c := range s.NextProposerQueue {
		s2.NextProposerQueue[i] = c
	}

	s2.CurrentEpochVoteAssignments = make([]uint32, len(s.CurrentEpochVoteAssignments))
	for i, c := range s.CurrentEpochVoteAssignments {
		s2.CurrentEpochVoteAssignments[i] = c
	}

	s2.PreviousEpochVoteAssignments = make([]uint32, len(s.PreviousEpochVoteAssignments))
	for i, c := range s.PreviousEpochVoteAssignments {
		s2.PreviousEpochVoteAssignments[i] = c
	}

	s2.LatestBlockHashes = make([]chainhash.Hash, len(s.LatestBlockHashes))
	for i, c := range s.LatestBlockHashes {
		s2.LatestBlockHashes[i] = c
	}

	s2.CurrentEpochVotes = make([]AcceptedVoteInfo, len(s.CurrentEpochVotes))
	for i, c := range s.CurrentEpochVotes {
		s2.CurrentEpochVotes[i] = c.Copy()
	}

	s2.PreviousEpochVotes = make([]AcceptedVoteInfo, len(s.PreviousEpochVotes))
	for i, c := range s.PreviousEpochVotes {
		s2.PreviousEpochVotes[i] = c.Copy()
	}

	s2.CurrentManagers = make([][20]byte, len(s.CurrentManagers))
	copy(s2.CurrentManagers, s.CurrentManagers)

	s2.ManagerReplacement = s.ManagerReplacement.Copy()

	s2.ReplaceVotes = make(map[[20]byte]chainhash.Hash, len(s.ReplaceVotes))
	for i, k := range s.ReplaceVotes {
		s2.ReplaceVotes[i] = k
	}

	s2.CommunityVotes = make(map[chainhash.Hash]CommunityVoteData, len(s.CommunityVotes))
	for i, k := range s.CommunityVotes {
		val := k.Copy()
		s2.CommunityVotes[i] = *val
	}

	return s2
}

// AccountTxs is just a helper struct for database storage of account transactions.
type AccountTxs struct {
	Amount uint64
	Txs    []chainhash.Hash
}

// Marshal encodes the data.
func (ac *AccountTxs) Marshal() ([]byte, error) {
	return ssz.Marshal(ac)
}

// Unmarshal decodes the data.
func (ac *AccountTxs) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, ac)
}
