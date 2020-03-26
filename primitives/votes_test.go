package primitives

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
)

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

	buf := bytes.NewBuffer([]byte{})

	if err := av.Serialize(buf); err != nil {
		t.Fatal(err)
	}

	av2 := &AcceptedVoteInfo{}
	if err := av2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(av, av2); diff != nil {
		t.Fatal(diff)
	}
}
