package state

// VoteData is the part of a vote that needs to be signed.
type VoteData struct {
	// Slot is the slot the validators were assigned.
	Slot uint64

	// FromEpoch is the source epoch of the vote which should either be
	// the current justified epoch or the previous justified epoch.
	FromEpoch uint64

	// FromHash is the block hash of the FromEpoch.
	FromHash []byte `ssz-size:"32"`

	// ToEpoch is the target epoch of the vote which should either be the
	// current epoch or the previous epoch.
	ToEpoch uint64

	// ToHash is the block hash of the ToEpoch.
	ToHash []byte `ssz-size:"32"`

	// BeaconBlockHash is for the fork choice.
	BeaconBlockHash []byte `ssz-size:"32"`
}

// AcceptedVoteInfo is vote data and participation for accepted votes.
type AcceptedVoteInfo struct {
	// Data is the data of the vote which specifies the signed part of the attestation.
	Data *VoteData

	// ParticipationBitfield is any validator that participated in the
	// vote.
	ParticipationBitfield []uint8 `ssz-max:"1099511627776"`

	// Proposer is the proposer that included the attestation in a block.
	Proposer uint32

	// InclusionDelay is the delay from the attestation slot to the slot
	// included.
	InclusionDelay uint64
}

// MaxVoteDataSize is the maximum size in bytes of vote data.
const MaxVoteDataSize = 8 + 8 + 32 + 8 + 32 + 32

// SingleValidatorVote is a signed vote from a validator.
type SingleValidatorVote struct {
	Data      *VoteData
	Signature []byte `ssz-size:"96"`
	Offset    uint32
	OutOf     uint32
}

// MultiValidatorVote is a vote signed by one or many validators.
type MultiValidatorVote struct {
	Data                  *VoteData
	Signature             []byte  `ssz-size:"96"`
	ParticipationBitfield []uint8 `ssz-max:"1099511627776"`
}
