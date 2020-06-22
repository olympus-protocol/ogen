package primitives_test

import (
	"errors"
	"testing"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/prysmaticlabs/go-ssz"
)

var voteData = primitives.VoteData{
	Slot:      1,
	FromEpoch: 2,
	FromHash:  [32]byte{3},
	ToEpoch:   4,
	ToHash:    [32]byte{5},
}

var acceptedVoteInfo = primitives.AcceptedVoteInfo{
	Data:                  voteData,
	ParticipationBitfield: []uint8{6, 7},
	Proposer:              8,
	InclusionDelay:        9,
}

var singleValidatorVote = primitives.SingleValidatorVote{
	Data:      voteData,
	Signature: bls.NewAggregateSignature().Marshal(),
	Offset:    10,
	OutOf:     15,
}

var multiValidatorVote = primitives.MultiValidatorVote{
	Data:                  voteData,
	Signature:             bls.NewAggregateSignature().Marshal(),
	ParticipationBitfield: []byte{1, 2, 3, 4},
}

func Test_VotesSerialize(t *testing.T) {
	err := serVoteData()
	if err != nil {
		t.Error(err)
	}
	err = serAcceptedVoteInfo()
	if err != nil {
		t.Error(err)
	}
	err = serSingleValidatorVote()
	if err != nil {
		t.Error(err)
	}
	err = serMultiValidatorVote()
	if err != nil {
		t.Error(err)
	}
}

func serVoteData() error {
	ser, err := voteData.Marshal()
	if err != nil {
		return err
	}
	var desc primitives.VoteData
	err = desc.Unmarshal(ser)
	if err != nil {
		return err
	}
	equal := ssz.DeepEqual(desc, voteData)
	if !equal {
		return errors.New("marshal/unmarshal failed for VoteData")
	}
	return nil
}

func serAcceptedVoteInfo() error {
	ser, err := acceptedVoteInfo.Marshal()
	if err != nil {
		return err
	}
	var des primitives.AcceptedVoteInfo
	err = des.Unmarshal(ser)
	if err != nil {
		return err
	}
	equal := ssz.DeepEqual(des, acceptedVoteInfo)
	if !equal {
		return errors.New("marshal/unmarshal failed for AcceptedVoteInfo")
	}
	return nil
}

func serSingleValidatorVote() error {
	ser, err := singleValidatorVote.Marshal()
	if err != nil {
		return err
	}
	var des primitives.SingleValidatorVote
	err = des.Unmarshal(ser)
	if err != nil {
		return err
	}
	equal := ssz.DeepEqual(des, singleValidatorVote)
	if !equal {
		return errors.New("marshal/unmarshal failed for SingleValidatorVote")
	}
	return nil
}

func serMultiValidatorVote() error {
	ser, err := multiValidatorVote.Marshal()
	if err != nil {
		return err
	}
	var des primitives.MultiValidatorVote
	err = des.Unmarshal(ser)
	if err != nil {
		return err
	}
	equal := ssz.DeepEqual(des, multiValidatorVote)
	if !equal {
		return errors.New("marshal/unmarshal failed for MultiValidatorVoteInfo")
	}
	return nil
}

func TestAcceptedVoteInfoCopy(t *testing.T) {
	av2 := acceptedVoteInfo.Copy()
	acceptedVoteInfo.Data.Slot = 2
	if av2.Data.Slot == 2 {
		t.Fatal("mutating data mutates copy")
	}

	acceptedVoteInfo.ParticipationBitfield[0] = 7
	if av2.ParticipationBitfield[0] == 7 {
		t.Fatal("mutating participation bitfield mutates copy")
	}

	acceptedVoteInfo.Proposer = 7
	if av2.Proposer == 7 {
		t.Fatal("mutating proposer mutates copy")
	}

	acceptedVoteInfo.InclusionDelay = 7
	if av2.InclusionDelay == 7 {
		t.Fatal("mutating inclusion delay mutates copy")
	}
}
