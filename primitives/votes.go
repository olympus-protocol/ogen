package primitives

import (
	"fmt"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
	"github.com/prysmaticlabs/go-ssz"

	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// AcceptedVoteInfo is vote data and participation for accepted votes.
type AcceptedVoteInfo struct {
	// Data is the data of the vote which specifies the signed part of the
	// attestation.
	Data VoteData

	// ParticipationBitfield is any validator that participated in the
	// vote.
	ParticipationBitfield bitfield.Bitfield

	// Proposer is the proposer that included the attestation in a block.
	Proposer uint32

	// InclusionDelay is the delay from the attestation slot to the slot
	// included.
	InclusionDelay uint64
}

// Marshal encodes the data.
func (av *AcceptedVoteInfo) Marshal() ([]byte, error) {
	return ssz.Marshal(av)
}

// Unmarshal decodes the data.
func (av *AcceptedVoteInfo) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, av)
}

// Copy returns a copy of the AcceptedVoteInfo.
func (a *AcceptedVoteInfo) Copy() AcceptedVoteInfo {
	a2 := *a

	a2.ParticipationBitfield = make([]uint8, len(a.ParticipationBitfield))
	for i, b := range a.ParticipationBitfield {
		a2.ParticipationBitfield[i] = b
	}

	a2.Data = a.Data.Copy()

	return a2
}

// MaxVoteDataSize is the maximum size in bytes of vote data.
const MaxVoteDataSize = 8 + 8 + 32 + 8 + 32 + 32

// VoteData is the part of a vote that needs to be signed.
type VoteData struct {
	// Slot is the slot the validators were assigned.
	Slot uint64

	// FromEpoch is the source epoch of the vote which should either be
	// the current justified epoch or the previous justified epoch.
	FromEpoch uint64

	// FromHash is the block hash of the FromEpoch.
	FromHash chainhash.Hash

	// ToEpoch is the target epoch of the vote which should either be the
	// current epoch or the previous epoch.
	ToEpoch uint64

	// ToHash is the block hash of the ToEpoch.
	ToHash chainhash.Hash

	// BeaconBlockHash is for the fork choice.
	BeaconBlockHash chainhash.Hash
}

// Marshal encodes the data.
func (v *VoteData) Marshal() ([]byte, error) {
	return ssz.Marshal(v)
}

// Unmarshal decodes the data.
func (v *VoteData) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, v)
}

func (v *VoteData) FirstSlotValid(p *params.ChainParams) uint64 {
	// vs <= ss-min
	// vs + min <= ss
	// state slot >= vote slot + min

	return v.Slot + p.MinAttestationInclusionDelay
}

func (v *VoteData) LastSlotValid(p *params.ChainParams) uint64 {
	// ss <= vs+epoch-1

	return v.Slot + p.EpochLength - 1
}

func (v *VoteData) String() string {
	return fmt.Sprintf("Vote(epochs: %d -> %d, beacon: %s)", v.FromEpoch, v.ToEpoch, v.BeaconBlockHash)
}

// IsDoubleVote checks if the two votes form a double vote.
func (v *VoteData) IsDoubleVote(v2 VoteData) bool {
	return v.ToEpoch == v2.ToEpoch && !v.Equals(&v2)
}

// IsSurroundVote checks if the two votes form a surrounded vote.
func (v *VoteData) IsSurroundVote(v2 VoteData) bool {
	return v.FromEpoch < v2.FromEpoch && v2.ToEpoch < v.ToEpoch
}

// Equals checks if vote data equals another vote data.
func (v *VoteData) Equals(other *VoteData) bool {
	if v.Slot != other.Slot || v.FromEpoch != other.FromEpoch || v.ToEpoch != other.ToEpoch ||
		!v.FromHash.IsEqual(&other.FromHash) || !v.ToHash.IsEqual(&other.ToHash) || !v.BeaconBlockHash.IsEqual(&other.BeaconBlockHash) {
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
	hash, _ := ssz.HashTreeRoot(v)
	return chainhash.Hash(hash)
}

// SingleValidatorVote is a signed vote from a validator.
type SingleValidatorVote struct {
	Data   VoteData
	Sig    []byte
	Offset uint32
	OutOf  uint32
}

// Signature returns the signature on BLS type
func (v *SingleValidatorVote) Signature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(v.Sig)
}

// Marshal encodes the data.
func (v *SingleValidatorVote) Marshal() ([]byte, error) {
	return ssz.Marshal(v)
}

// Unmarshal decodes the data.
func (v *SingleValidatorVote) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, v)
}

// AsMulti returns the single validator vote as a multi validator vote.
func (v *SingleValidatorVote) AsMulti() *MultiValidatorVote {
	participationBitfield := make([]uint8, (v.OutOf+7)/8)
	participationBitfield[v.Offset/8] |= (1 << uint(v.Offset%8))
	return &MultiValidatorVote{
		Data:                  v.Data,
		Sig:                   v.Sig,
		ParticipationBitfield: participationBitfield,
	}
}

func (v *SingleValidatorVote) Hash() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(v)
	return chainhash.Hash(hash)
}

// MultiValidatorVote is a vote signed by one or many validators.
type MultiValidatorVote struct {
	Data                  VoteData
	Sig                   []byte
	ParticipationBitfield bitfield.Bitfield
}

// Signature returns the signature on BLS type
func (v *MultiValidatorVote) Signature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(v.Sig)
}

// Marshal encodes the data.
func (v *MultiValidatorVote) Marshal() ([]byte, error) {
	return ssz.Marshal(v)
}

// Unmarshal decodes the data.
func (v *MultiValidatorVote) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, v)
}

// Hash calculates the hash of the vote.
func (v *MultiValidatorVote) Hash() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(v)
	return chainhash.Hash(hash)
}
