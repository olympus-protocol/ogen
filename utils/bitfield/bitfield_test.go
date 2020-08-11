package bitfield_test

import (
	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Bitlist(t *testing.T) {
	bf := bitfield.NewBitlist(4 * 8)

	bf.Set(32)

	assert.True(t, bf.Get(32))

	assert.Equal(t, bf.Len(), uint64(32))
}

func TestBitlist_Intersect(t *testing.T) {
	a := bitfield.NewBitlist(50000)
	b := bitfield.NewBitlist(50000)

	a.Set(10)
	a.Set(5)
	b.Set(5)
	b.Set(3)

	intersect := a.Intersect(b)
	assert.Equal(t, intersect, []int{5})
}
