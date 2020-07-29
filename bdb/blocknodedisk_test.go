package bdb_test

import (
	"testing"

	"github.com/olympus-protocol/ogen/bdb"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
)

func Test_BlockNodeDiskSerialize(t *testing.T) {
	ser, err := testdata.BlockNode.Marshal()

	assert.NoError(t, err)

	var desc bdb.BlockNodeDisk

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, testdata.BlockNode, desc)
}
