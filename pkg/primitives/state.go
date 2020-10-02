package primitives

import "github.com/olympus-protocol/ogen/pkg/bitfield"

const MaxStateSize = 1024 * 1024 * 2.5 // 2.5 MB

// SerializableState is a serializable copy of the state
type SerializableState struct {
	// CoinsState keeps if accounts balances and transactions.
	CoinsState *CoinsStateSerializable

	// ValidatorRegistry keeps track of validators in the state.
	ValidatorRegistry []*Validator `ssz-max:"2097152"`

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
	CurrentEpochVotes []*AcceptedVoteInfo `ssz-max:"2097152"`

	// PreviousJustifiedEpoch is the second-to-last epoch that >2/3 of validators
	// voted for.
	PreviousJustifiedEpoch uint64

	// PreviousJustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	PreviousJustifiedEpochHash [32]byte `ssz-size:"32"`

	// PreviousEpochVotes are votes where the FromEpoch matches PreviousJustifiedEpoch.
	PreviousEpochVotes []*AcceptedVoteInfo `ssz-max:"2097152"`

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
