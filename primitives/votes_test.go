package primitives

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/olympus-protocol/ogen/bls"
)

func TestMultiValidatorVoteSerializeDeserialize(t *testing.T) {
	var s [96]byte
	sig := bls.NewAggregateSignature()
	copy(s[:], sig.Marshal())
	bl := &MultiValidatorVote{
		Data: VoteData{
			Slot:      1,
			FromEpoch: 2,
			FromHash:  [32]byte{3},
			ToEpoch:   4,
			ToHash:    [32]byte{5},
		},
		Sig:                   s,
		ParticipationBitfield: []byte{1, 2, 3, 4},
	}

	buf, err := bl.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	av2 := &MultiValidatorVote{}
	if err := av2.Unmarshal(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(bl, av2); diff != nil {
		t.Fatal(diff)
	}
}

func TestAcceptedVoteInfoCopy(t *testing.T) {
	av := &AcceptedVoteInfo{
		Data: VoteData{
			Slot:      1,
			FromEpoch: 2,
			FromHash:  [32]byte{3},
			ToEpoch:   4,
			ToHash:    [32]byte{5},
		},
		ParticipationBitfield: []uint8{6, 7},
		Proposer:              8,
		InclusionDelay:        9,
	}

	av2 := av.Copy()
	av.Data.Slot = 2
	if av2.Data.Slot == 2 {
		t.Fatal("mutating data mutates copy")
	}

	av.ParticipationBitfield[0] = 7
	if av2.ParticipationBitfield[0] == 7 {
		t.Fatal("mutating participation bitfield mutates copy")
	}

	av.Proposer = 7
	if av2.Proposer == 7 {
		t.Fatal("mutating proposer mutates copy")
	}

	av.InclusionDelay = 7
	if av2.InclusionDelay == 7 {
		t.Fatal("mutating inclusion delay mutates copy")
	}
}

func TestAcceptedVoteInfoSerializeDeserialize(t *testing.T) {
	av := &AcceptedVoteInfo{
		Data: VoteData{
			Slot:      1,
			FromEpoch: 2,
			FromHash:  [32]byte{3},
			ToEpoch:   4,
			ToHash:    [32]byte{5},
		},
		ParticipationBitfield: []uint8{6, 7},
		Proposer:              8,
		InclusionDelay:        9,
	}

	buf, err := av.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	av2 := &AcceptedVoteInfo{}
	if err := av2.Unmarshal(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(av, av2); diff != nil {
		t.Fatal(diff)
	}
}
