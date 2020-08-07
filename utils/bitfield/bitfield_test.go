package bitfield_test

import (
	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Bitfield(t *testing.T) {
	bf := bitfield.NewBitlist(4 * 8)

	bf.Set(32)

	assert.True(t, bf.Get(32))

	assert.Equal(t, bf.Len(), uint64(32))
}
