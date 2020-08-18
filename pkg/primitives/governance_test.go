package primitives_test

import (
	"github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/bls/multisig"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CommunityVoteDataSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 5)
	var v primitives.CommunityVoteData
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.CommunityVoteData
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_ReplacementVoteSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.ReplacementVotes
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.ReplacementVotes
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_CommunityVoteSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 5)
	var v primitives.CommunityVoteData
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.CommunityVoteData
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_GovernanceVoteSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.GovernanceVote
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.GovernanceVote
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func TestGovernanceVote_Copy(t *testing.T) {
	v := primitives.GovernanceVote{
		Type: 1,
		Data: [100]byte{1, 2, 3},
		CombinedSig: &multisig.CombinedSignature{
			S: [96]byte{1, 2, 3},
			P: [48]byte{1, 2, 3},
		},
		VoteEpoch: 1,
	}

	v2 := v.Copy()

	v.Type = 2
	assert.Equal(t, v2.Type, uint64(1))

	v.Data[0] = 2
	assert.Equal(t, v2.Data[0], uint8(1))

	v.CombinedSig.S[0] = 2
	assert.Equal(t, v2.CombinedSig.S[0], uint8(1))

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

// Is not possible to test against equal states because of slice ordering. TODO find a solution
func Test_GovernanceSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 5)

	replace := map[[20]byte]chainhash.Hash{}
	community := map[chainhash.Hash]primitives.CommunityVoteData{}
	f.Fuzz(&replace)
	f.Fuzz(&community)

	v := primitives.Governance{
		ReplaceVotes:   replace,
		CommunityVotes: community,
	}

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Governance
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)
}

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
