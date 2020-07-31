package conflict_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/peers/conflict"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ValidatorHelloMessageSerializtion(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v conflict.ValidatorHelloMessage
	f.Fuzz(&v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	var desc conflict.ValidatorHelloMessage

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

