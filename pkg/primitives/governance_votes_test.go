package primitives_test

import (
	"encoding/hex"
	"github.com/olympus-protocol/ogen/pkg/bls"
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

	assert.Equal(t, "0af37aec3b567cb1c5465c229e438f10b981a685cc44b39ad70d783ec76e5574", orig.Hash().String())
}

func TestGovernanceVote(t *testing.T) {
	marshal := testdata.FuzzGovernanceVote(10)
	for _, c := range marshal {

		ser, err := c.Marshal()
		assert.NoError(t, err)

		assert.LessOrEqual(t, len(ser), primitives.MaxGovernanceVoteSize)

		desc := new(primitives.GovernanceVote)

		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)

		assert.True(t, c.Valid())
	}

	rawKey, err := hex.DecodeString("2f00e92c4fd8b0afb02f502a982971cfed1ef3efc92dea16a99a911c6bf1ce88")
	assert.NoError(t, err)

	k, err := bls.SecretKeyFromBytes(rawKey)
	assert.NoError(t, err)

	g := &primitives.GovernanceVote{
		Type:      10,
		Data:      []byte{1, 2, 3, 4, 5, 6},
		VoteEpoch: 10,
		Signature: [96]byte{},
		PublicKey: [48]byte{},
	}

	msg := g.SignatureHash()
	sig := k.Sign(msg[:])
	pub := k.PublicKey()

	copy(g.Signature[:], sig.Marshal())
	copy(g.PublicKey[:], pub.Marshal())

	cp := g.Copy()

	g.VoteEpoch = 11
	assert.Equal(t, uint64(10), cp.VoteEpoch)

	g.Type = 11
	assert.Equal(t, uint64(10), cp.Type)

	g.Data[1] = 5
	assert.Equal(t, uint8(2), cp.Data[1])

	assert.Equal(t, "166caeb545b78480be8b55542f2807bccf5938639963f3ef90a4cc0d73199b77", g.Hash().String())
}
