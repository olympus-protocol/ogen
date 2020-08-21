package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidator(t *testing.T) {
	v := testdata.FuzzValidator(10)
	for _, c := range v {
		ser, err := c.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.Validator)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)
	}

	orig := primitives.Validator{
		Balance:          100,
		PubKey:           [48]byte{1, 2, 3, 4, 5},
		PayeeAddress:     [20]byte{1, 2, 3, 4, 5},
		Status:           5,
		FirstActiveEpoch: 1,
		LastActiveEpoch:  10,
	}

	cp := orig.Copy()

	orig.Balance = 110
	assert.Equal(t, cp.Balance, uint64(100))

	orig.PubKey[47] = 10
	assert.Equal(t, cp.PubKey[47], uint8(0))

	orig.PayeeAddress[19] = 10
	assert.Equal(t, cp.PayeeAddress[19], uint8(0))

	orig.Status = 10
	assert.Equal(t, cp.Status, uint64(5))

	orig.FirstActiveEpoch = 10
	assert.Equal(t, cp.FirstActiveEpoch, uint64(1))

	orig.LastActiveEpoch = 15
	assert.Equal(t, cp.LastActiveEpoch, uint64(10))

	orig.Status = primitives.StatusActive
	assert.True(t, orig.IsActive())

	orig.Status = primitives.StatusActivePendingExit
	assert.True(t, orig.IsActive())

	orig.Status = primitives.StatusExitedWithPenalty
	assert.False(t, orig.IsActive())

	orig.Status = primitives.StatusExitedWithoutPenalty
	assert.Equal(t, "EXITED", orig.StatusString())
	orig.Status = primitives.StatusExitedWithPenalty
	assert.Equal(t, "PENALTY_EXIT", orig.StatusString())
	orig.Status = primitives.StatusActivePendingExit
	assert.Equal(t, "PENDING_EXIT", orig.StatusString())
	orig.Status = primitives.StatusActive
	assert.Equal(t, "ACTIVE", orig.StatusString())
	orig.Status = primitives.StatusStarting
	assert.Equal(t, "STARTING", orig.StatusString())

	orig.Status = primitives.StatusActive
	assert.True(t, orig.IsActiveAtEpoch(11))
	assert.False(t, orig.IsActiveAtEpoch(9))
}
