package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/google/gofuzz"
)

func Test_BlockNodeDiskSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.BlockNodeDisk
	f.Fuzz(&v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	var desc primitives.BlockNodeDisk

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}
