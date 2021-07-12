package primitives

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls/common"

	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/params"

	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// AcceptedVoteInfo is vote data and participation for accepted votes.
type AcceptedVoteInfo struct {
	// Data is the data of the vote which specifies the signed part of the attestation.
	Data *VoteData

	// ParticipationBitfield is any validator that participated in the vote.
	// Max size is the same as the MultiValidatorVote bitlist size.
	ParticipationBitfield bitfield.Bitlist `ssz:"bitlist" ssz-max:"6250"`

	// Proposer is the proposer that included the attestation in a block.
	Proposer uint64

	// InclusionDelay is the delay from the attestation slot to the slot included.
	InclusionDelay uint64
}

// Marshal encodes the data.
func (a *AcceptedVoteInfo) Marshal() ([]byte, error) {
	b, err := a.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Unmarshal decodes the data.
func (a *AcceptedVoteInfo) Unmarshal(b []byte) error {
	return a.UnmarshalSSZ(b)
}

// Copy returns a copy of the AcceptedVoteInfo.
func (a *AcceptedVoteInfo) Copy() AcceptedVoteInfo {
	a2 := *a

	a2.ParticipationBitfield = bitfield.NewBitlist(a.ParticipationBitfield.Len())
	for i, b := range a.ParticipationBitfield {
		a2.ParticipationBitfield[i] = b
	}

	vd := a.Data.Copy()
	a2.Data = &vd

	return a2
}

// VoteData is the part of a vote that needs to be signed.
type VoteData struct {
	// Slot is the slot the validators were assigned.
	Slot uint64

	// FromEpoch is the source epoch of the vote which should either be
	// the current justified epoch or the previous justified epoch.
	FromEpoch uint64

	// FromHash is the block hash of the FromEpoch.
	FromHash [32]byte

	// ToEpoch is the target epoch of the vote which should either be the
	// current epoch or the previous epoch.
	ToEpoch uint64

	// ToHash is the block hash of the ToEpoch.
	ToHash [32]byte

	// BeaconBlockHash is for the fork choice.
	BeaconBlockHash [32]byte
}

// Marshal encodes the data.
func (v *VoteData) Marshal() ([]byte, error) {
	b, err := v.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Unmarshal decodes the data.
func (v *VoteData) Unmarshal(b []byte) error {
	return v.UnmarshalSSZ(b)
}

// FirstSlotValid return the first slot valid for current validator vote
func (v *VoteData) FirstSlotValid(p *params.ChainParams) uint64 {
	// vs <= ss-min
	// vs + min <= ss
	// state slot >= vote slot + min
	return v.Slot + p.MinAttestationInclusionDelay
}

// LastSlotValid return the last slot valid for current validator vote
func (v *VoteData) LastSlotValid(p *params.ChainParams) uint64 {
	// ss <= vs+epoch-1
	return v.Slot + p.EpochLength - 1
}

func (v *VoteData) String() string {
	return fmt.Sprintf("Vote(epochs: %d -> %d, beacon: %s)", v.FromEpoch, v.ToEpoch, hex.EncodeToString(v.BeaconBlockHash[:]))
}

// IsDoubleVote checks if the two votes form a double vote.
func (v *VoteData) IsDoubleVote(v2 *VoteData) bool {
	return v.ToEpoch == v2.ToEpoch && !v.Equals(v2)
}

// IsSurroundVote checks if the two votes form a surrounded vote.
func (v *VoteData) IsSurroundVote(v2 *VoteData) bool {
	return v.FromEpoch < v2.FromEpoch && v2.ToEpoch < v.ToEpoch
}

// Equals checks if vote data equals another vote data.
func (v *VoteData) Equals(other *VoteData) bool {
	if v.Slot != other.Slot || v.FromEpoch != other.FromEpoch || v.ToEpoch != other.ToEpoch ||
		!bytes.Equal(v.FromHash[:], other.FromHash[:]) || !bytes.Equal(v.ToHash[:], other.ToHash[:]) || !bytes.Equal(v.BeaconBlockHash[:], other.BeaconBlockHash[:]) {
		return false
	}

	return true
}

// Copy returns a copy of the vote data.
func (v *VoteData) Copy() VoteData {
	return *v
}

// Hash calculates the hash of the vote data.
func (v *VoteData) Hash() chainhash.Hash {
	cp := v.Copy()
	b, _ := cp.Marshal()
	return chainhash.HashH(b)
}

// MultiValidatorVote is a vote signed by one or many validators.
type MultiValidatorVote struct {
	// Data defines the vote properties.
	Data *VoteData
	// Sig is the aggregated signature for all validators voting for the VoteData.
	Sig [96]byte
	// ParticipationBitfield is a bitlist that marks the index of the validators voting.
	// Maximum amount of votes inside a bitlist is 50000
	ParticipationBitfield bitfield.Bitlist `ssz:"bitlist" ssz-max:"50000"`
}

// Signature returns the signature on BLS type
func (m *MultiValidatorVote) Signature() (common.Signature, error) {
	return bls.SignatureFromBytes(m.Sig[:])
}

// Marshal encodes the data.
func (m *MultiValidatorVote) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal decodes the data.
func (m *MultiValidatorVote) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}
