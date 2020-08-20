package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ValidatorHelloMessageSerializtion(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.ValidatorHelloMessage
	f.Fuzz(&v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	var desc primitives.ValidatorHelloMessage

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}
