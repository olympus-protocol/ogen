package primitives_test

import (
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGovernanceVote_Copy(t *testing.T) {
	v := primitives.GovernanceVote{
		Type:          1,
		Data:          [100]byte{1, 2, 3},
		FunctionalSig: [144]byte{1, 2, 3},
		VoteEpoch:     1,
	}

	v2 := v.Copy()

	v.Type = 2
	assert.Equal(t, v2.Type, uint64(1))

	v.Data[0] = 2
	assert.Equal(t, v2.Data[0], uint8(1))

	v.FunctionalSig[0] = 2
	assert.Equal(t, v2.FunctionalSig[0], uint8(1))

	v.VoteEpoch = 2
	assert.Equal(t, v2.VoteEpoch, uint64(1))
}

func TestCommunityVoteData_Copy(t *testing.T) {
	v := primitives.CommunityVoteData{
		ReplacementCandidates: [][20]byte{
			{1, 2, 3},
			{1, 2, 3},
			{1, 2, 3},
			{1, 2, 3},
			{1, 2, 3},
		},
	}

	v2 := v.Copy()

	v.ReplacementCandidates[0][0] = 2
	assert.Equal(t, v2.ReplacementCandidates[0][0], uint8(1))
}
