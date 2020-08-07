package bitfield

import (
	"math/bits"
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

// NewBitlist creates a new bitlist of size N.
func NewBitlist(n uint64) Bitlist {
	ret := make(Bitlist, n/8+1)

	// Set most significant bit for length bit.
	i := uint8(1 << (n % 8))
	ret[n/8] |= i

	return ret
}
