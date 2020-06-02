package primitives

import (
	"bytes"
	"fmt"
	"github.com/olympus-protocol/ogen/params"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// AcceptedVoteInfo is vote data and participation for accepted votes.
type AcceptedVoteInfo struct {
	// Data is the data of the vote which specifies the signed part of the
	// attestation.
	Data VoteData

	// ParticipationBitfield is any validator that participated in the
	// vote.
	ParticipationBitfield []uint8

	// Proposer is the proposer that included the attestation in a block.
	Proposer uint32

	// InclusionDelay is the delay from the attestation slot to the slot
	// included.
	InclusionDelay uint64
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

// Serialize serializes the accepted vote info to a writer.
func (a *AcceptedVoteInfo) Serialize(w io.Writer) error {
	if err := a.Data.Serialize(w); err != nil {
		return err
	}
	if err := serializer.WriteVarBytes(w, a.ParticipationBitfield); err != nil {
		return err
	}
	return serializer.WriteElements(w, a.Proposer, a.InclusionDelay)
}

// Deserialize deserializes the accepted vote info from a reader.
func (a *AcceptedVoteInfo) Deserialize(r io.Reader) (err error) {
	if err := a.Data.Deserialize(r); err != nil {
		return err
	}
	a.ParticipationBitfield, err = serializer.ReadVarBytes(r)
	return serializer.ReadElements(r, &a.Proposer, &a.InclusionDelay)
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
	buf := bytes.NewBuffer([]byte{})
	_ = v.Serialize(buf)
	return chainhash.HashH(buf.Bytes())
}

// Serialize serializes the vote data to a writer.
func (v *VoteData) Serialize(w io.Writer) error {
	return serializer.WriteElements(w, v.Slot, v.FromEpoch, v.FromHash, v.ToEpoch, v.ToHash, v.BeaconBlockHash)
}

// Deserialize deserializes the vote data from a reader.
func (v *VoteData) Deserialize(r io.Reader) error {
	return serializer.ReadElements(r, &v.Slot, &v.FromEpoch, &v.FromHash, &v.ToEpoch, &v.ToHash, &v.BeaconBlockHash)
}

// SingleValidatorVote is a signed vote from a validator.
type SingleValidatorVote struct {
	Data      VoteData
	Signature bls.Signature
	Offset    uint32
	OutOf     uint32
}

// AsMulti returns the single validator vote as a multi validator vote.
func (v *SingleValidatorVote) AsMulti() *MultiValidatorVote {
	participationBitfield := make([]uint8, (v.OutOf+7)/8)
	participationBitfield[v.Offset/8] |= (1 << uint(v.Offset%8))
	return &MultiValidatorVote{
		Data:                  v.Data,
		Signature:             v.Signature,
		ParticipationBitfield: participationBitfield,
	}
}

func (v *SingleValidatorVote) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = v.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}

// Serialize serializes a SingleValidatorVote to a writer.
func (v *SingleValidatorVote) Encode(w io.Writer) error {
	if err := v.Data.Serialize(w); err != nil {
		return err
	}
	sig := v.Signature.Marshal()
	if _, err := w.Write(sig[:]); err != nil {
		return err
	}
	return serializer.WriteElements(w, v.Offset, v.OutOf)
}

// Deserialize deserializes a SingleValidatorVote from a reader.
func (v *SingleValidatorVote) Decode(r io.Reader) error {
	if err := v.Data.Deserialize(r); err != nil {
		return err
	}
	sigBytes := make([]byte, 96)
	if _, err := r.Read(sigBytes[:]); err != nil {
		return err
	}
	sig, err := bls.SignatureFromBytes(sigBytes)
	if err != nil {
		return err
	}
	v.Signature = *sig
	return serializer.ReadElements(r, &v.Offset, &v.OutOf)
}

// MultiValidatorVote is a vote signed by one or many validators.
type MultiValidatorVote struct {
	Data                  VoteData
	Signature             bls.Signature
	ParticipationBitfield []uint8
}

// Hash calculates the hash of the vote.
func (v *MultiValidatorVote) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = v.Serialize(buf)
	return chainhash.HashH(buf.Bytes())
}

// Serialize serializes a MultiValidatorVote to a writer.
func (v *MultiValidatorVote) Serialize(w io.Writer) error {
	if err := v.Data.Serialize(w); err != nil {
		return err
	}
	sig := v.Signature.Marshal()
	if err := serializer.WriteElement(w, sig); err != nil {
		return err
	}
	return serializer.WriteVarBytes(w, v.ParticipationBitfield)
}

// Deserialize deserializes a MultiValidatorVote from a reader.
func (v *MultiValidatorVote) Deserialize(r io.Reader) (err error) {
	if err := v.Data.Deserialize(r); err != nil {
		return err
	}
	sigBytes := make([]byte, 96)
	if err := serializer.ReadElement(r, sigBytes[:]); err != nil {
		return err
	}
	sig, err := bls.SignatureFromBytes(sigBytes)
	if err != nil {
		return err
	}

	v.Signature = *sig
	v.ParticipationBitfield, err = serializer.ReadVarBytes(r)
	return
}
