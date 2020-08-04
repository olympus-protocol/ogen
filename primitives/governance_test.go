package primitives_test

import (
	"github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
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

func TestGovernance_Copy(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 5)

	replace := map[[20]byte]chainhash.Hash{}
	community := map[chainhash.Hash]primitives.CommunityVoteData{}
	f.Fuzz(&replace)
	f.Fuzz(&community)

	keyRep := [20]byte{1, 2, 3}
	keyCom := [32]byte{1, 2, 3}

	g := primitives.Governance{
		ReplaceVotes:   replace,
		CommunityVotes: community,
	}

	g.ReplaceVotes[keyRep] = [32]byte{1, 2, 3, 4}
	g.CommunityVotes[keyCom] = primitives.CommunityVoteData{
		ReplacementCandidates: [][20]byte{
			{1, 2, 3},
		},
	}

	g2 := g.Copy()

	g.ReplaceVotes[keyRep] = [32]byte{1, 2, 3, 5}
	g2RepData := g2.ReplaceVotes[keyRep]
	expBytes := [32]byte{1, 2, 3, 4}
	assert.Equal(t, g2RepData[:], expBytes[:])

	g.CommunityVotes[keyCom] = primitives.CommunityVoteData{
		ReplacementCandidates: [][20]byte{
			{1, 2, 5},
		},
	}
	assert.Equal(t, g2.CommunityVotes[keyCom].ReplacementCandidates, [][20]byte{{1, 2, 3}})
}
