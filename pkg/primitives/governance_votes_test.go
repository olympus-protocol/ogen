package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/multisig"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReplacementVotes(t *testing.T) {
	v := testdata.FuzzReplacementVote(10)

	for _, c := range v {
		ser, err := c.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.ReplacementVotes)

		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)
	}

}

func TestCommunityVoteData(t *testing.T) {
	v := testdata.FuzzCommunityVoteData(10)

	for _, c := range v {
		ser, err := c.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.CommunityVoteData)

		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)
	}

	orig := primitives.CommunityVoteData{
		ReplacementCandidates: [][20]byte{
			{1, 2, 3},
			{1, 2, 3},
			{1, 2, 3},
			{1, 2, 3},
			{1, 2, 3},
		},
	}

	cp := orig.Copy()

	orig.ReplacementCandidates[0][0] = 2
	assert.Equal(t, cp.ReplacementCandidates[0][0], uint8(1))

	assert.Equal(t, "74556ec73e780dd79ab344cc85a681b9108f439e225c46c5b17c563bec7af30a", orig.Hash().String())
}

func TestGovernanceVote(t *testing.T) {
	marshal := testdata.FuzzGovernanceVote(10)
	for _, c := range marshal {

		ser, err := c.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.GovernanceVote)

		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)

		assert.True(t, c.Valid())
	}

	g := &primitives.GovernanceVote{
		Type:      10,
		Data:      [100]byte{1, 2, 3, 4, 5, 6},
		Multisig:  nil,
		VoteEpoch: 10,
	}

	mp := multisig.NewMultipub([]*bls.PublicKey{}, 5)
	ms := multisig.NewMultisig(mp)

	g.Multisig = ms

	cp := g.Copy()

	g.VoteEpoch = 11
	assert.Equal(t, uint64(10), cp.VoteEpoch)

	g.Type = 11
	assert.Equal(t, uint64(10), cp.Type)

	g.Data[1] = 5
	assert.Equal(t, uint8(2), cp.Data[1])

	assert.Equal(t, "e2692dfd1afebb607da9e57b00d7bf67d05e15ef22b9915832fafa692c546b6a", g.Hash().String())
}
