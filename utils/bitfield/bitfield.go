package bitfcheck

import "github.com/prysmaticlabs/go-bitfield"

func Set(b bitfield.Bitlist, i uint) {
	b[i/8] |= 1 << (i % 8)
	return
}

func Get(b bitfield.Bitlist, i uint) bool {
	if b[i/8]&(1<<(i%8)) != 0 {
		return true
	}
	return false
}
