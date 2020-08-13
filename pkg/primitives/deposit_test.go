package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_DepositSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.Deposit
	f.Fuzz(&v)
	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Deposit
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_DeposiDatatSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.DepositData
	f.Fuzz(&v)
	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.DepositData
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}
