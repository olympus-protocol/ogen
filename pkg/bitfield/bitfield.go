package bitfield

import (
	"math/bits"
	"reflect"
)

type Bitlist []byte

// Len of the bitlist returns the number of bits available in the underlying byte array.
func (b Bitlist) Len() uint64 {
	if len(b) == 0 {
		return 0
	}
	// The most significant bit is present in the last byte in the array.
	last := b[len(b)-1]

	// Determine the position of the most significant bit.
	msb := bits.Len8(last)

	// The absolute position of the most significant bit will be the number of
	// bits in the preceding bytes plus the position of the most significant
	// bit. Subtract this value by 1 to determine the length of the bitlist.
	return uint64(8*(len(b)-1) + msb - 1)
}

// Set marks the specified bit
func (b Bitlist) Set(i uint) {
	b[i/8] |= 1 << (i % 8)
	return
}

// Get returns the specified bit
func (b Bitlist) Get(i uint) bool {
	if b[i/8]&(1<<(i%8)) != 0 {
		return true
	}
	return false
}

// Overlaps returns true if the bitlist contains one of the bits from the provided argument
// bitlist. This method will panic if bitlists are not the same length.
func (b Bitlist) Overlaps(c Bitlist) bool {
	lenB, lenC := b.Len(), c.Len()
	if lenB != lenC {
		panic("bitlists are different lengths")
	}

	if lenB == 0 || lenC == 0 {
		return false
	}

	msb := uint8(bits.Len8(b[len(b)-1])) - 1
	lengthBitMask := uint8(1 << msb)

	// To ensure all of the bits in c are not overlapped in b, we iterate over every byte, invert b
	// and xor the byte from b and c, then and it against c. If the result is non-zero, then
	// we can be assured that byte in c had bits not overlapped in b.
	for i := 0; i < len(b); i++ {
		// If this byte is the last byte in the array, mask the length bit.
		mask := uint8(0xFF)
		if i == len(b)-1 {
			mask &^= lengthBitMask
		}

		if (^b[i]^c[i])&c[i]&mask != 0 {
			return true
		}
	}
	return false
}

// Contains returns true if the bitlist contains all of the bits from the provided argument
// bitlist. This method will panic if bitlists are not the same length.
func (b Bitlist) Contains(c Bitlist) bool {
	if b.Len() != c.Len() {
		panic("bitlists are different lengths")
	}

	// To ensure all of the bits in c are present in b, we iterate over every byte, combine
	// the byte from b and c, then XOR them against b. If the result of this is non-zero, then we
	// are assured that a byte in c had bits not present in b.
	for i := 0; i < len(b); i++ {
		if b[i]^(b[i]|c[i]) != 0 {
			return false
		}
	}

	return true
}

// Count returns the number of 1s in the bitlist.
func (b Bitlist) Count() uint64 {
	c := 0

	for _, bt := range b {
		c += bits.OnesCount8(bt)
	}

	if c > 0 {
		c-- // Remove length bit from count.
	}

	return uint64(c)
}

// Intersect returns the bit indices of intersection between two bitlists
func (b Bitlist) Intersect(c Bitlist) []int {
	a := b.BitIndices()
	e := c.BitIndices()

	set := make([]int, 0)

	for i := 0; i < len(a); i++ {
		if contains(e, a[i]) {
			set = append(set, a[i])
		}
	}

	return set
}

// BitIndices returns an slice of int with the indexes marked on the bitlist
func (b Bitlist) BitIndices() []int {
	indices := make([]int, 0, b.Count())
	for i, bt := range b {
		if i == len(b)-1 {
			// Clear the most significant bit (the length bit).
			msb := uint8(bits.Len8(bt)) - 1
			bt &^= uint8(1 << msb)
		}
		for j := 0; j < 8; j++ {
			bit := byte(1 << uint(j))
			if bt&bit == bit {
				indices = append(indices, i*8+j)
			}
		}
	}

	return indices
}

// NewBitlist creates a new bitlist of size N.
func NewBitlist(n uint64) Bitlist {
	ret := make(Bitlist, n/8+1)

	// Set most significant bit for length bit.
	i := uint8(1 << (n % 8))
	ret[n/8] |= i

	return ret
}

func contains(a interface{}, e interface{}) bool {
	v := reflect.ValueOf(a)
	for i := 0; i < v.Len(); i++ {
		if v.Index(i).Interface() == e {
			return true
		}
	}
	return false
}
