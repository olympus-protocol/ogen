package bdb_test

import (
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/bdb"
	"github.com/stretchr/testify/assert"
)

func Test_BlockNodeDiskSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)

	v := new(bdb.BlockNodeDisk)
	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(bdb.BlockNodeDisk)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}
