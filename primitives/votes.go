package primitives

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	bitfcheck "github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/prysmaticlabs/go-bitfield"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"

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
	MaxVoteDataSize = 128
	// MaxAcceptedVoteInfoSize is the maximum size in bytes an accepted vote info can contain.
	MaxAcceptedVoteInfoSize = MaxVoteDataSize + 2064
	// MaxSingleValidatorVoteSize is the maximum size in bytes a single validator vote can contain.
	MaxSingleValidatorVoteSize = MaxVoteDataSize + 112
	// MaxMultiValidatorVoteSize is the maximum size in bytes a multi validator vote can contain.
	MaxMultiValidatorVoteSize = MaxVoteDataSize + 2144
)

// AcceptedVoteInfo is vote data and participation for accepted votes.
type AcceptedVoteInfo struct {
	// Data is the data of the vote which specifies the signed part of the attestation.
	Data *VoteData

	// ParticipationBitfield is any validator that participated in the vote.
	ParticipationBitfield bitfield.Bitlist `ssz:"bitlist" ssz-max:"2048"`

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
	if len(b) > MaxAcceptedVoteInfoSize {
		return nil, ErrorAcceptedVoteDataSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (a *AcceptedVoteInfo) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxAcceptedVoteInfoSize {
		return ErrorAcceptedVoteDataSize
	}
	return a.UnmarshalSSZ(d)
}

// Copy returns a copy of the AcceptedVoteInfo.
func (a *AcceptedVoteInfo) Copy() AcceptedVoteInfo {
	a2 := *a

	a2.ParticipationBitfield = bitfield.NewBitlist(a.ParticipationBitfield.Len())
	for i, b := range a.ParticipationBitfield.Bytes() {
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

	// Nonce identifies the client that proposed the block.
	Nonce uint64
}

// Marshal encodes the data.
func (v *VoteData) Marshal() ([]byte, error) {
	b, err := v.MarshalSSZ()
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
	return v.UnmarshalSSZ(d)
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
		!bytes.Equal(v.FromHash[:], other.FromHash[:]) || bytes.Equal(v.ToHash[:], other.ToHash[:]) || bytes.Equal(v.BeaconBlockHash[:], other.BeaconBlockHash[:]) || v.Nonce != other.Nonce {
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
	Data   *VoteData
	Sig    [96]byte
	Offset uint64
	OutOf  uint64
}

// Signature returns the signature on BLS type
func (s *SingleValidatorVote) Signature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(s.Sig)
}

// Marshal encodes the data.
func (s *SingleValidatorVote) Marshal() ([]byte, error) {
	b, err := s.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxSingleValidatorVoteSize {
		return nil, ErrorSingleValidatorVoteSize
	}
	return b, nil
}

// Unmarshal decodes the data.
func (s *SingleValidatorVote) Unmarshal(b []byte) error {
	if len(b) > MaxSingleValidatorVoteSize {
		return ErrorSingleValidatorVoteSize
	}
	return s.UnmarshalSSZ(b)
}

// AsMulti returns the single validator vote as a multi validator vote.
func (s *SingleValidatorVote) AsMulti() *MultiValidatorVote {
	participationBitfield := bitfield.NewBitlist(s.OutOf + 7)
	bitfcheck.Set(participationBitfield, uint(s.Offset))
	return &MultiValidatorVote{
		Data:                  s.Data,
		Sig:                   s.Sig,
		ParticipationBitfield: participationBitfield,
	}
}

// Hash returns the hash of the single validator vote.
func (s *SingleValidatorVote) Hash() chainhash.Hash {
	b, _ := s.Marshal()
	return chainhash.HashH(b)
}

// MultiValidatorVote is a vote signed by one or many validators.
type MultiValidatorVote struct {
	Data                  *VoteData
	Sig                   [96]byte
	ParticipationBitfield bitfield.Bitlist `ssz:"bitlist" ssz-max:"2048"`
}

// Signature returns the signature on BLS type
func (m *MultiValidatorVote) Signature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(m.Sig)
}

// Marshal encodes the data.
func (m *MultiValidatorVote) Marshal() ([]byte, error) {
	b, err := m.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxMultiValidatorVoteSize {
		return nil, ErrorMultiValidatorVoteSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (m *MultiValidatorVote) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxMultiValidatorVoteSize {
		return ErrorMultiValidatorVoteSize
	}
	return m.UnmarshalSSZ(d)
}

// Hash calculates the hash of the vote.
func (m *MultiValidatorVote) Hash() chainhash.Hash {
	b, _ := m.Marshal()
	return chainhash.HashH(b)
}
