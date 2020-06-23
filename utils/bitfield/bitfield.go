package bitfield

// Bitfield is a bitfield of a certain length.
type Bitfield struct {
	Field []byte `ssz-max:"1024"`
}

// NewBitfield constructs a new bitfield containing a certain length.
func NewBitfield(l uint) *Bitfield {
	return &Bitfield{Field: make([]byte, (l+7)/8)}
}

// Set sets bit i
func (b Bitfield) Set(i uint) {
	b.Field[i/8] |= (1 << (i % 8))
}

// Get gets bit i
func (b Bitfield) Get(i uint) bool {
	return b.Field[i/8]&(1<<(i%8)) != 0
}

// MaxLength is the maximum number of elements the bitfield can hold.
func (b Bitfield) MaxLength() uint {
	return uint(len(b.Field)) * 8
}

// Copy returns a copy of the bitfield.
func (b Bitfield) Copy() Bitfield {
	newB := Bitfield{Field: (make([]byte, len(b.Field)))}
	copy(newB.Field, b.Field)
	return newB
}
