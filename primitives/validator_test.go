package primitives_test

import (
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidator_Copy(t *testing.T) {
	v := primitives.Validator{
		Balance:          100,
		PubKey:           [48]byte{1, 2, 3, 4, 5},
		PayeeAddress:     [20]byte{1, 2, 3, 4, 5},
		Status:           5,
		FirstActiveEpoch: 1,
		LastActiveEpoch:  10,
	}

	v2 := v.Copy()

	v.Balance = 110
	assert.Equal(t, v2.Balance, uint64(100))

	v.PubKey[47] = 10
	assert.Equal(t, v2.PubKey[47], uint8(0))

	v.PayeeAddress[19] = 10
	assert.Equal(t, v2.PayeeAddress[19], uint8(0))

	v.Status = 10
	assert.Equal(t, v2.Status, uint64(5))

	v.FirstActiveEpoch = 10
	assert.Equal(t, v2.FirstActiveEpoch, uint64(1))

	v.LastActiveEpoch = 15
	assert.Equal(t, v2.LastActiveEpoch, uint64(10))
}
