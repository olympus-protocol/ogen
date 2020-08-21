package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGovernance_Copy(t *testing.T) {
	g := testdata.FuzzGovernanceState()

	keyRep := [20]byte{1, 2, 3}
	keyCom := [32]byte{1, 2, 3}

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

// Is not possible to test against equal states because of slice ordering. TODO find a solution
func Test_GovernanceSerialize(t *testing.T) {
	v := testdata.FuzzGovernanceState()

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Governance
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)
}

func TestGovernance_ToSerializable(t *testing.T) {
	g := testdata.FuzzGovernanceState()

	sgs := g.ToSerializable()

	assert.Equal(t, len(g.ReplaceVotes), len(sgs.ReplaceVotes))
	assert.Equal(t, len(g.CommunityVotes), len(sgs.CommunityVotes))
}

func TestGovernance_FromSerializable(t *testing.T) {
	sgs := testdata.FuzzGovernanceStateSerializable()

	gs := new(primitives.Governance)
	gs.FromSerializable(sgs)

	assert.Equal(t, len(gs.ReplaceVotes), len(sgs.ReplaceVotes))
	assert.Equal(t, len(gs.CommunityVotes), len(sgs.CommunityVotes))
}
