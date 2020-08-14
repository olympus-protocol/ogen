package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CoinStateSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(10000, 10000)
	balances := map[[20]byte]uint64{}
	nonces := map[[20]byte]uint64{}
	f.Fuzz(&balances)
	f.Fuzz(&nonces)

	v := primitives.CoinsState{
		Balances: balances,
		Nonces:   nonces,
	}

	ser, err := v.Marshal()

	assert.NoError(t, err)

	var desc primitives.CoinsState
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func TestCoinsState_ToSerializable(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 20)

	var balances map[[20]byte]uint64
	var nonces map[[20]byte]uint64

	f.Fuzz(&balances)
	f.Fuzz(&nonces)

	cs := primitives.CoinsState{
		Balances: balances,
		Nonces:   nonces,
	}

	scs := cs.ToSerializable()

	assert.Equal(t, len(scs.Nonces), len(nonces))
	assert.Equal(t, len(scs.Balances), len(balances))

	noncesSumMap := uint64(0)
	balancesSumMap := uint64(0)

	for _, v := range balances {
		balancesSumMap += v
	}

	for _, v := range nonces {
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
}

func TestCoinsState_FromSerializable(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 20)

	var balances []*primitives.AccountInfo
	var nonces []*primitives.AccountInfo

	f.Fuzz(&balances)
	f.Fuzz(&nonces)

	scs := &primitives.CoinsStateSerializable{
		Balances: balances,
		Nonces:   nonces,
	}

	cs := new(primitives.CoinsState)
	cs.FromSerializable(scs)

	assert.Equal(t, len(cs.Nonces), len(nonces))
	assert.Equal(t, len(cs.Balances), len(balances))

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

	for _, v := range nonces {
		noncesSumSlice += v.Info
	}

	for _, v := range balances {
		balancesSumSlice += v.Info
	}

	assert.Equal(t, noncesSumMap, noncesSumSlice)
	assert.Equal(t, balancesSumMap, balancesSumSlice)
}

func TestCoinsState_Copy(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 20)

	var balances map[[20]byte]uint64
	var nonces map[[20]byte]uint64

	key := [20]byte{1, 2, 3}

	f.Fuzz(&balances)
	f.Fuzz(&nonces)

	balances[key] = 10
	nonces[key] = 10

	cs := primitives.CoinsState{
		Balances: balances,
		Nonces:   nonces,
	}

	cs2 := cs.Copy()

	cs.Nonces[key] = 11
	cs.Balances[key] = 11

	assert.Equal(t, cs2.Nonces[key], uint64(10))
	assert.Equal(t, cs2.Balances[key], uint64(10))
}
