package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGovernance_ToSerializable(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 20)

	var replace map[[20]byte]chainhash.Hash
	var community map[chainhash.Hash]primitives.CommunityVoteData

	f.Fuzz(&replace)
	f.Fuzz(&community)

	gs := primitives.Governance{
		ReplaceVotes:   replace,
		CommunityVotes: community,
	}

	sgs := gs.ToSerializable()

	assert.Equal(t, len(gs.ReplaceVotes), len(sgs.ReplaceVotes))
	assert.Equal(t, len(gs.CommunityVotes), len(sgs.CommunityVotes))
}

func TestGovernance_FromSerializable(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 20)

	var replace []*primitives.ReplacementVotes
	var community []*primitives.CommunityVoteDataInfo

	f.Fuzz(&replace)
	f.Fuzz(&community)

	sgs := &primitives.GovernanceSerializable{
		ReplaceVotes:   replace,
		CommunityVotes: community,
	}

	gs := new(primitives.Governance)
	gs.FromSerializable(sgs)

	assert.Equal(t, len(gs.ReplaceVotes), len(sgs.ReplaceVotes))
	assert.Equal(t, len(gs.CommunityVotes), len(sgs.CommunityVotes))
}
