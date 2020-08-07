package primitives_test

import (
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVoteData_Copy(t *testing.T) {
	v := &primitives.VoteData{
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
	assert.Equal(t, v2.Slot, uint64(5))

	v.FromEpoch = 6
	assert.Equal(t, v2.FromEpoch, uint64(5))

	v.FromHash[31] = 10
	assert.Equal(t, v2.FromHash[31], uint8(0))

	v.ToEpoch = 10
	assert.Equal(t, v2.ToEpoch, uint64(5))

	v.ToHash[31] = 10
	assert.Equal(t, v2.ToHash[31], uint8(0))

	v.BeaconBlockHash[31] = 10
	assert.Equal(t, v2.BeaconBlockHash[31], uint8(0))

	v.Nonce = 10
	assert.Equal(t, v2.Nonce, uint64(5))
}

func TestAcceptedVoteInfo_Copy(t *testing.T) {
	av := &primitives.AcceptedVoteInfo{
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

	av2 := av.Copy()

	av.Data.Slot = 2
	assert.Equal(t, av2.Data.Slot, uint64(1))

	assert.Equal(t, av.ParticipationBitfield.Len(), av2.ParticipationBitfield.Len())

	assert.Equal(t, len(av.ParticipationBitfield), len(av2.ParticipationBitfield))

	assert.Equal(t, av.ParticipationBitfield, av2.ParticipationBitfield)

	av.ParticipationBitfield.Set(uint(2))

	assert.Equal(t, av2.ParticipationBitfield[0], uint8(0))

	av.Proposer = 7

	assert.Equal(t, av2.Proposer, uint64(8))

	av.InclusionDelay = 7

	assert.Equal(t, av2.InclusionDelay, uint64(9))

}

func TestSingleValidatorVote_AsMulti(t *testing.T) {
	s := &primitives.SingleValidatorVote{
		Data: &primitives.VoteData{
			Slot:            5,
			FromEpoch:       5,
			FromHash:        [32]byte{1, 2, 3},
			ToEpoch:         5,
			ToHash:          [32]byte{1, 2, 3},
			BeaconBlockHash: [32]byte{1, 2, 3},
			Nonce:           5,
		},
		Sig:    [96]byte{1, 2, 3},
		Offset: 56,
		OutOf:  512,
	}

	mv := s.AsMulti()

	assert.Equal(t, mv.Data, s.Data)

	assert.Equal(t, mv.Sig, s.Sig)

	b, err := mv.Marshal()
	assert.NoError(t, err)

	var m primitives.MultiValidatorVote

	err = m.Unmarshal(b)

	assert.NoError(t, err)

	assert.Equal(t, mv, &m)

}
