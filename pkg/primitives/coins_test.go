package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CoinStateSerialize(t *testing.T) {
	v := fuzzCoinState(10)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.CoinsState)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func TestCoinsState_ToSerializable(t *testing.T) {
	cs := fuzzCoinState(10)

	scs := cs.ToSerializable()

	assert.Equal(t, len(scs.Nonces), 10)
	assert.Equal(t, len(scs.Balances), 10)

	noncesSumMap := uint64(0)
	balancesSumMap := uint64(0)

	for _, v := range cs.Balances {
		balancesSumMap += v
	}

	for _, v := range cs.Nonces {
		noncesSumMap += v
	}

	noncesSumSlice := uint64(0)
	balancesSumSlice := uint64(0)

	for _, v := range scs.Nonces {
		noncesSumSlice += v.Info
	}

	for _, v := range scs.Balances {
		balancesSumSlice += v.Info
	}

	assert.Equal(t, noncesSumMap, noncesSumSlice)
	assert.Equal(t, balancesSumMap, balancesSumSlice)

	assert.Equal(t, balancesSumMap, cs.GetTotal())
}

func TestCoinsState_FromSerializable(t *testing.T) {
	scs := fuzzCoinStateSerializable(10)

	cs := new(primitives.CoinsState)
	cs.FromSerializable(scs)

	assert.Equal(t, len(cs.Nonces), 10)
	assert.Equal(t, len(cs.Balances), 10)

	noncesSumMap := uint64(0)
	balancesSumMap := uint64(0)

	for _, v := range cs.Balances {
		balancesSumMap += v
	}

	for _, v := range cs.Nonces {
		noncesSumMap += v
	}

	noncesSumSlice := uint64(0)
	balancesSumSlice := uint64(0)

	for _, v := range scs.Nonces {
		noncesSumSlice += v.Info
	}

	for _, v := range scs.Balances {
		balancesSumSlice += v.Info
	}

	assert.Equal(t, noncesSumMap, noncesSumSlice)
	assert.Equal(t, balancesSumMap, balancesSumSlice)

	assert.Equal(t, balancesSumMap, cs.GetTotal())
}

func TestCoinsState_Copy(t *testing.T) {
	cs := fuzzCoinState(10)
	key := [20]byte{1, 2, 3}

	cs.Balances[key] = 10
	cs.Nonces[key] = 10

	cs2 := cs.Copy()

	cs.Nonces[key] = 11
	cs.Balances[key] = 11

	assert.Equal(t, cs2.Nonces[key], uint64(10))
	assert.Equal(t, cs2.Balances[key], uint64(10))
}
