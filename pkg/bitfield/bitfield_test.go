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

func TestBitlist_Contains(t *testing.T) {
	a := bitfield.NewBitlist(50000)
	b := bitfield.NewBitlist(50000)
	c := bitfield.NewBitlist(10)

	a.Set(10)
	b.Set(10)

	contains, err := a.Contains(b)
	assert.NoError(t, err)
	assert.True(t, contains)

	_, err = a.Contains(c)
	assert.Equal(t, bitfield.ErrorsBitlistSize, err)
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
	c := bitfield.NewBitlist(50000)

	a.Set(10)
	a.Set(5)
	b.Set(300)
	b.Set(700)
	c.Set(10)

	merge, err := a.Merge(b)
	assert.NoError(t, err)
	assert.True(t, merge.Get(uint(10)))
	assert.True(t, merge.Get(uint(5)))
	assert.True(t, merge.Get(uint(300)))
	assert.True(t, merge.Get(uint(700)))

	_, err = a.Merge(c)
	assert.Equal(t, bitfield.ErrorBitlistOverlaps, err)
}
