package bitfield_test

import (
	"github.com/olympus-protocol/ogen/pkg/bitfield"
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

func TestBitlist_Merge(t *testing.T) {
	a := bitfield.NewBitlist(50000)
	b := bitfield.NewBitlist(50000)

	a.Set(10)
	a.Set(5)
	b.Set(300)
	b.Set(700)

	merge, err := a.Merge(b)
	assert.NoError(t, err)
	assert.True(t, merge.Get(uint(10)))
	assert.True(t, merge.Get(uint(5)))
	assert.True(t, merge.Get(uint(300)))
	assert.True(t, merge.Get(uint(700)))

}
