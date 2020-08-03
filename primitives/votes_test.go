package primitives

import (
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAcceptedVoteInfoCopy(t *testing.T) {
	av := &AcceptedVoteInfo{
		Data: &VoteData{
			Slot:      1,
			FromEpoch: 2,
			FromHash:  [32]byte{3},
			ToEpoch:   4,
			ToHash:    [32]byte{5},
		},
		ParticipationBitfield: bitfield.NewBitlist(2 * 8),
		Proposer:              8,
		InclusionDelay:        9,
	}

	av2 := av.Copy()
	av.Data.Slot = 2
	assert.NotEqual(t, av2.Data.Slot, 2)

	av.ParticipationBitfield[0] = 7

	assert.NotEqual(t, av2.ParticipationBitfield[0], 7)

	av.Proposer = 7

	assert.NotEqual(t, av2.Proposer, 7)

	av.InclusionDelay = 7

	assert.NotEqual(t, av2.InclusionDelay, 7)
}
