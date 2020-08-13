package blockdb_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/internal/blockdb"
)

func Test_BlockNodeDiskSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v blockdb.BlockNodeDisk
	f.Fuzz(&v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	var desc blockdb.BlockNodeDisk

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}
