package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAcceptedVoteInfo(t *testing.T) {

	// Test correct AcceptedVoteInfo
	correct := fuzzAcceptedVoteInfo(10, true, true)

	for _, c := range correct {
		data, err := c.Marshal()
		assert.NoError(t, err)

		n := new(primitives.AcceptedVoteInfo)
		err = n.Unmarshal(data)
		assert.NoError(t, err)

		assert.Equal(t, c, n)
	}

	// Test wrong sized data
	incorrect := fuzzAcceptedVoteInfo(10, false, true)

	for _, c := range incorrect {
		_, err := c.Marshal()
		assert.NotNil(t, err)
	}

	// Test marshal/unmarshal not panic accessing a nil pointer
	// Should create all data to null values
	nildata := fuzzAcceptedVoteInfo(10, true, false)

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

	c := fuzzAcceptedVoteInfo(1, true, true)
	original := c[0]
	cp := original.Copy()

	// Test that both elements have same information
	assert.Equal(t, original.InclusionDelay, cp.InclusionDelay)
	assert.Equal(t, original.Proposer, cp.Proposer)

	contains, err := original.ParticipationBitfield.Contains(cp.ParticipationBitfield)
	assert.NoError(t, err)
	assert.True(t, contains)

	assert.Equal(t, original.Data, cp.Data)

	// Check original and copy hold separated values

}

//func Test_AcceptedVoteInfoSerialize(t *testing.T) {
//	f := fuzz.New().NilChance(0)
//	var v primitives.AcceptedVoteInfo
//	f.Fuzz(&v)
//	v.ParticipationBitfield = bitfield.NewBitlist(8)
//
//	ser, err := v.Marshal()
//	assert.NoError(t, err)
//
//	var desc primitives.AcceptedVoteInfo
//	err = desc.Unmarshal(ser)
//	assert.NoError(t, err)
//
//	assert.Equal(t, v, desc)
//}
//
//func Test_VoteDataSerialize(t *testing.T) {
//	f := fuzz.New().NilChance(0)
//	var v primitives.VoteData
//	f.Fuzz(&v)
//
//	ser, err := v.Marshal()
//	assert.NoError(t, err)
//
//	var desc primitives.VoteData
//	err = desc.Unmarshal(ser)
//	assert.NoError(t, err)
//
//	assert.Equal(t, v, desc)
//}
//
//func Test_MultiValidatorVoteSerialize(t *testing.T) {
//	v := fuzzMultiValidatorVote(1)
//	ser, err := v[0].Marshal()
//	assert.NoError(t, err)
//
//	desc := new(primitives.MultiValidatorVote)
//	err = desc.Unmarshal(ser)
//	assert.NoError(t, err)
//
//	assert.Equal(t, v[0], desc)
//}
//
//func TestVoteData_Copy(t *testing.T) {
//	v := &primitives.VoteData{
//		Slot:            5,
//		FromEpoch:       5,
//		FromHash:        [32]byte{1, 2, 3},
//		ToEpoch:         5,
//		ToHash:          [32]byte{1, 2, 3},
//		BeaconBlockHash: [32]byte{1, 2, 3},
//		Nonce:           5,
//	}
//
//	v2 := v.Copy()
//
//	v.Slot = 6
//	assert.Equal(t, v2.Slot, uint64(5))
//
//	v.FromEpoch = 6
//	assert.Equal(t, v2.FromEpoch, uint64(5))
//
//	v.FromHash[31] = 10
//	assert.Equal(t, v2.FromHash[31], uint8(0))
//
//	v.ToEpoch = 10
//	assert.Equal(t, v2.ToEpoch, uint64(5))
//
//	v.ToHash[31] = 10
//	assert.Equal(t, v2.ToHash[31], uint8(0))
//
//	v.BeaconBlockHash[31] = 10
//	assert.Equal(t, v2.BeaconBlockHash[31], uint8(0))
//
//	v.Nonce = 10
//	assert.Equal(t, v2.Nonce, uint64(5))
//}
//
//func TestAcceptedVoteInfo_Copy(t *testing.T) {
//	av := &primitives.AcceptedVoteInfo{
//		Data: &primitives.VoteData{
//			Slot:      1,
//			FromEpoch: 2,
//			FromHash:  [32]byte{3},
//			ToEpoch:   4,
//			ToHash:    [32]byte{5},
//		},
//		ParticipationBitfield: bitfield.NewBitlist(16),
//		Proposer:              8,
//		InclusionDelay:        9,
//	}
//
//	av2 := av.Copy()
//
//	av.Data.Slot = 2
//	assert.Equal(t, av2.Data.Slot, uint64(1))
//
//	assert.Equal(t, av.ParticipationBitfield.Len(), av2.ParticipationBitfield.Len())
//
//	assert.Equal(t, len(av.ParticipationBitfield), len(av2.ParticipationBitfield))
//
//	assert.Equal(t, av.ParticipationBitfield, av2.ParticipationBitfield)
//
//	av.ParticipationBitfield.Set(uint(2))
//
//	assert.Equal(t, av2.ParticipationBitfield[0], uint8(0))
//
//	av.Proposer = 7
//
//	assert.Equal(t, av2.Proposer, uint64(8))
//
//	av.InclusionDelay = 7
//
//	assert.Equal(t, av2.InclusionDelay, uint64(9))
//
//}
