package primitives

import (
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVoteData_Copy(t *testing.T) {
	v := &VoteData{
		Slot:            5,
		FromEpoch:       5,
		FromHash:        [32]byte{1, 2, 3},
		ToEpoch:         5,
		ToHash:          [32]byte{1, 2, 3},
		BeaconBlockHash: [32]byte{1, 2, 3},
		Nonce:           5,
	}

	v2 := v.Copy()

	v.Slot = 6
	assert.NotEqual(t, v2.Slot, 6)

	v.FromEpoch = 6
	assert.NotEqual(t, v2.FromEpoch, 6)

	v.FromHash[31] = 10
	assert.NotEqual(t, v2.FromHash[31], 10)

	v.ToEpoch = 10
	assert.NotEqual(t, v2.ToEpoch, 10)

	v.ToHash[31] = 10
	assert.NotEqual(t, v2.ToHash[31], 10)

	v.BeaconBlockHash[31] = 10
	assert.NotEqual(t, v2.BeaconBlockHash[31], 10)

	v.Nonce = 10
	assert.NotEqual(t, v2.Nonce, 10)
}

func TestAcceptedVoteInfo_Copy(t *testing.T) {
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
