package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVoteData(t *testing.T) {
	v := testdata.FuzzVoteData(10)
	for _, c := range v {
		ser, err := c.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.VoteData)

		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)
	}

	d := &primitives.VoteData{
		Slot:            5,
		FromEpoch:       5,
		FromHash:        [32]byte{1, 2, 3},
		ToEpoch:         5,
		ToHash:          [32]byte{1, 2, 3},
		BeaconBlockHash: [32]byte{1, 2, 3},
		Nonce:           5,
	}

	dv := &primitives.VoteData{
		Slot:            5,
		FromEpoch:       5,
		FromHash:        [32]byte{1, 2, 3},
		ToEpoch:         5,
		ToHash:          [32]byte{1, 2, 3},
		BeaconBlockHash: [32]byte{1, 2, 4},
		Nonce:           5,
	}

	sv := &primitives.VoteData{
		Slot:            5,
		FromEpoch:       6,
		FromHash:        [32]byte{1, 2, 3},
		ToEpoch:         4,
		ToHash:          [32]byte{1, 2, 3},
		BeaconBlockHash: [32]byte{1, 2, 3},
		Nonce:           5,
	}

	assert.Equal(t, "Vote(epochs: 5 -> 5, beacon: 0102030000000000000000000000000000000000000000000000000000000000)", d.String())
	assert.Equal(t, uint64(6), d.FirstSlotValid(&testdata.TestParams))
	assert.Equal(t, uint64(9), d.LastSlotValid(&testdata.TestParams))
	assert.True(t, d.Equals(d))
	assert.Equal(t, "1f2fffad96474211ab3b06df31c9657557ec38bc39e0a5aa15e98ac8dc0bdba6", d.Hash().String())
	assert.False(t, d.IsDoubleVote(d))
	assert.False(t, d.IsSurroundVote(d))
	assert.True(t, d.IsDoubleVote(dv))
	assert.True(t, d.IsSurroundVote(sv))

	cp := d.Copy()

	d.Slot = 6
	assert.Equal(t, cp.Slot, uint64(5))

	d.FromEpoch = 6
	assert.Equal(t, cp.FromEpoch, uint64(5))

	d.FromHash[31] = 10
	assert.Equal(t, cp.FromHash[31], uint8(0))

	d.ToEpoch = 10
	assert.Equal(t, cp.ToEpoch, uint64(5))

	d.ToHash[31] = 10
	assert.Equal(t, cp.ToHash[31], uint8(0))

	d.BeaconBlockHash[31] = 10
	assert.Equal(t, cp.BeaconBlockHash[31], uint8(0))

	d.Nonce = 10
	assert.Equal(t, cp.Nonce, uint64(5))
}

func TestAcceptedVoteInfo(t *testing.T) {

	// Test correct AcceptedVoteInfo
	correct := testdata.FuzzAcceptedVoteInfo(10, true, true)

	for _, c := range correct {
		data, err := c.Marshal()
		assert.NoError(t, err)

		n := new(primitives.AcceptedVoteInfo)
		err = n.Unmarshal(data)
		assert.NoError(t, err)

		assert.Equal(t, c, n)
	}

	// Test wrong sized data
	incorrect := testdata.FuzzAcceptedVoteInfo(10, false, true)

	for _, c := range incorrect {
		_, err := c.Marshal()
		assert.NotNil(t, err)
	}

	// Test marshal/unmarshal not panic accessing a nil pointer
	// Should create all data to null values
	nildata := testdata.FuzzAcceptedVoteInfo(10, true, false)

	for _, c := range nildata {
		assert.NotPanics(t, func() {
			data, err := c.Marshal()
			assert.NoError(t, err)

			n := new(primitives.AcceptedVoteInfo)
			err = n.Unmarshal(data)
			assert.NoError(t, err)

			assert.Equal(t, c, n)

			assert.Equal(t, uint64(0), n.Data.Slot)
			assert.Equal(t, uint64(0), n.Data.Nonce)
			assert.Equal(t, uint64(0), n.Data.FromEpoch)
			assert.Equal(t, uint64(0), n.Data.ToEpoch)
			assert.Equal(t, [32]byte{}, n.Data.BeaconBlockHash)
			assert.Equal(t, [32]byte{}, n.Data.FromHash)
			assert.Equal(t, [32]byte{}, n.Data.ToHash)
		})
	}

	orig := &primitives.AcceptedVoteInfo{
		Data: &primitives.VoteData{
			Slot:      1,
			FromEpoch: 2,
			FromHash:  [32]byte{3},
			ToEpoch:   4,
			ToHash:    [32]byte{5},
		},
		ParticipationBitfield: bitfield.NewBitlist(16),
		Proposer:              8,
		InclusionDelay:        9,
	}

	cp := orig.Copy()

	// Test that both elements have same information
	assert.Equal(t, orig.InclusionDelay, cp.InclusionDelay)
	assert.Equal(t, orig.Proposer, cp.Proposer)

	contains, err := orig.ParticipationBitfield.Contains(cp.ParticipationBitfield)
	assert.NoError(t, err)
	assert.True(t, contains)

	assert.Equal(t, orig.Data, cp.Data)

	orig.Data.Slot = 2
	assert.Equal(t, cp.Data.Slot, uint64(1))

	orig.ParticipationBitfield.Set(uint(2))

	assert.False(t, cp.ParticipationBitfield.Get(2))

	orig.Proposer = 7
	assert.Equal(t, cp.Proposer, uint64(8))

	orig.InclusionDelay = 7
	assert.Equal(t, cp.InclusionDelay, uint64(9))

}

func TestMultiValidatorVote(t *testing.T) {
	// Test correct AcceptedVoteInfo
	correct := testdata.FuzzMultiValidatorVote(10, true, true)

	for _, c := range correct {
		data, err := c.Marshal()
		assert.NoError(t, err)

		n := new(primitives.MultiValidatorVote)
		err = n.Unmarshal(data)
		assert.NoError(t, err)

		assert.Equal(t, c, n)
	}

	// Test wrong sized data
	incorrect := testdata.FuzzMultiValidatorVote(10, false, true)

	for _, c := range incorrect {
		_, err := c.Marshal()
		assert.NotNil(t, err)
	}

	// Test marshal/unmarshal not panic accessing a nil pointer
	// Should create all data to null values
	nildata := testdata.FuzzMultiValidatorVote(10, true, false)

	for _, c := range nildata {
		assert.NotPanics(t, func() {
			data, err := c.Marshal()
			assert.NoError(t, err)

			n := new(primitives.MultiValidatorVote)
			err = n.Unmarshal(data)
			assert.NoError(t, err)

			assert.Equal(t, c, n)

			assert.Equal(t, uint64(0), n.Data.Slot)
			assert.Equal(t, uint64(0), n.Data.Nonce)
			assert.Equal(t, uint64(0), n.Data.FromEpoch)
			assert.Equal(t, uint64(0), n.Data.ToEpoch)
			assert.Equal(t, [32]byte{}, n.Data.BeaconBlockHash)
			assert.Equal(t, [32]byte{}, n.Data.FromHash)
			assert.Equal(t, [32]byte{}, n.Data.ToHash)
		})
	}

	d := &primitives.MultiValidatorVote{
		Data: &primitives.VoteData{
			Slot:            5,
			FromEpoch:       5,
			FromHash:        [32]byte{1, 2, 3},
			ToEpoch:         5,
			ToHash:          [32]byte{1, 2, 3},
			BeaconBlockHash: [32]byte{1, 2, 3},
			Nonce:           5,
		},
		ParticipationBitfield: bitfield.NewBitlist(6242),
	}
	var sig [96]byte
	copy(sig[:], bls.NewAggregateSignature().Marshal())
	d.Sig = sig

	assert.Equal(t, "1f2fffad96474211ab3b06df31c9657557ec38bc39e0a5aa15e98ac8dc0bdba6", d.Data.Hash().String())
	newSig, err := d.Signature()
	assert.NoError(t, err)
	assert.NotNil(t, newSig)
}
