package bitfield

import (
	"io"

	"github.com/olympus-protocol/ogen/utils/serializer"
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

// Encode encodes the bitfield to the writer.
func (b Bitfield) Encode(w io.Writer) error {
	return serializer.WriteVarBytes(w, b)
}

// Decode decodes the bitfield from a reader.
func (b *Bitfield) Decode(r io.Reader) (err error) {
	*b, err = serializer.ReadVarBytes(r)
	return err
}

// Copy returns a copy of the bitfield.
func (b Bitfield) Copy() Bitfield {
	newB := Bitfield(make([]byte, len(b)))
	copy(newB, b)
	return newB
}
