package primitives

import (
	"errors"
	"fmt"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
	"github.com/prysmaticlabs/go-ssz"

	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var (
	// ErrorVoteDataSize is returned when a vote data is above MaxVoteDataSize
	ErrorVoteDataSize = errors.New("vote data too big")
	// ErrorAcceptedVoteDataSize is returned when a vote data is above MaxAcceptedVoteInfoSize
	ErrorAcceptedVoteDataSize = errors.New("accepted vote data too big")
	// ErrorSingleValidatorVoteSize is returned when a single validator vote data is above MaxSingleValidatorVoteSize
	ErrorSingleValidatorVoteSize = errors.New("single validator vote data too big")
	// ErrorMultiValidatorVoteSize is returned when a multi validator vote data is above MaxMultiValidatorVoteSize
	ErrorMultiValidatorVoteSize = errors.New("accepted vote data too big")
)

const (
	// MaxVoteDataSize is the maximum size in bytes of vote data.
	MaxVoteDataSize = 120
	// MaxAcceptedVoteInfoSize is the maximum size in bytes an accepted vote info can contain.
	MaxAcceptedVoteInfoSize = 140
	// MaxSingleValidatorVoteSize is the maximum size in bytes a single validator vote can contain.
	MaxSingleValidatorVoteSize = 228
	// MaxMultiValidatorVoteSize is the maximum size in bytes a multi validator vote can contain.
	MaxMultiValidatorVoteSize = 228
)

// AcceptedVoteInfo is vote data and participation for accepted votes.
type AcceptedVoteInfo struct {
	// Data is the data of the vote which specifies the signed part of the attestation.
	Data VoteData

	// ParticipationBitfield is any validator that participated in the vote.
	ParticipationBitfield bitfield.Bitfield

	// Proposer is the proposer that included the attestation in a block.
	Proposer uint32

	// InclusionDelay is the delay from the attestation slot to the slot included.
	InclusionDelay uint64
}

// Marshal encodes the data.
func (av *AcceptedVoteInfo) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(av)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxAcceptedVoteInfoSize {
		return nil, ErrorAcceptedVoteDataSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (av *AcceptedVoteInfo) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxAcceptedVoteInfoSize {
		return ErrorAcceptedVoteDataSize
	}
	return ssz.Unmarshal(d, av)
}

// Copy returns a copy of the AcceptedVoteInfo.
func (av *AcceptedVoteInfo) Copy() AcceptedVoteInfo {
	a2 := *av

	a2.ParticipationBitfield = make([]uint8, len(av.ParticipationBitfield))
	for i, b := range av.ParticipationBitfield {
		a2.ParticipationBitfield[i] = b
	}

	a2.Data = av.Data.Copy()

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
	b, err := ssz.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxVoteDataSize {
		return nil, ErrorVoteDataSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (v *VoteData) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxVoteDataSize {
		return ErrorVoteDataSize
	}
	return ssz.Unmarshal(d, v)
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
	b, _ := v.Marshal()
	return chainhash.HashH(b)
}

// SingleValidatorVote is a signed vote from a validator.
type SingleValidatorVote struct {
	Data   VoteData
	Sig    [96]byte
	Offset uint32
	OutOf  uint32
}

// Signature returns the signature on BLS type
func (v *SingleValidatorVote) Signature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(v.Sig)
}

// Marshal encodes the data.
func (v *SingleValidatorVote) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxSingleValidatorVoteSize {
		return nil, ErrorSingleValidatorVoteSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (v *SingleValidatorVote) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxSingleValidatorVoteSize {
		return ErrorSingleValidatorVoteSize
	}
	return ssz.Unmarshal(d, v)
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

// Hash returns the hash of the single validator vote.
func (v *SingleValidatorVote) Hash() chainhash.Hash {
	b, _ := v.Marshal()
	return chainhash.HashH(b)
}

// MultiValidatorVote is a vote signed by one or many validators.
type MultiValidatorVote struct {
	Data                  VoteData
	Sig                   [96]byte
	ParticipationBitfield bitfield.Bitfield
}

// Signature returns the signature on BLS type
func (v *MultiValidatorVote) Signature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(v.Sig)
}

// Marshal encodes the data.
func (v *MultiValidatorVote) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxMultiValidatorVoteSize {
		return nil, ErrorMultiValidatorVoteSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (v *MultiValidatorVote) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxMultiValidatorVoteSize {
		return ErrorMultiValidatorVoteSize
	}
	return ssz.Unmarshal(d, v)
}

// Hash calculates the hash of the vote.
func (v *MultiValidatorVote) Hash() chainhash.Hash {
	b, _ := v.Marshal()
	return chainhash.HashH(b)
}
