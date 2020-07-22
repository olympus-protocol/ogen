package bitfield

import (
	"github.com/golang/snappy"
	"github.com/prysmaticlabs/go-ssz"
)

// Bitfield is a bitfield of a certain length.
type Bitfield []byte

// NewBitfield constructs a new bitfield containing a certain length.
func NewBitfield(l uint) Bitfield {
	return make([]byte, (l+7)/8)
}

// Set sets bit i
func (b Bitfield) Set(i uint) {
	b[i/8] |= (1 << (i % 8))
}

// Get gets bit i
func (b Bitfield) Get(i uint) bool {
	return b[i/8]&(1<<(i%8)) != 0
}

// MaxLength is the maximum number of elements the bitfield can hold.
func (b Bitfield) MaxLength() uint {
	return uint(len(b)) * 8
}

// Marshal encodes the data.
func (b Bitfield) Marshal() ([]byte, error) {
	by, err := ssz.Marshal(b)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, by), nil
}

// Unmarshal decodes the data.
func (b Bitfield) Unmarshal(by []byte) error {
	de, err := snappy.Decode(nil, by)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(de, b)
}

// Copy returns a copy of the bitfield.
func (b Bitfield) Copy() Bitfield {
	newB := Bitfield(make([]byte, len(b)))
	copy(newB, b)
	return newB
}
