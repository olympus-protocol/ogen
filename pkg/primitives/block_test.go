package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlock(t *testing.T) {
	correct := testdata.FuzzBlock(2, true, true)
	for _, b := range correct {
		ser, err := b.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.Block)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, b, desc)
	}

	incorrect := testdata.FuzzBlock(2, false, true)
	for _, b := range incorrect {
		_, err := b.Marshal()
		assert.NotNil(t, err)

	}

	nilpointers := testdata.FuzzBlock(2, true, false)
	for _, b := range nilpointers {
		assert.NotPanics(t, func() {
			data, err := b.Marshal()
			assert.NoError(t, err)

			n := new(primitives.Block)
			err = n.Unmarshal(data)
			assert.NoError(t, err)

			assert.Equal(t, b, n)

			assert.Equal(t, uint64(0), n.Header.Slot)
		})

	}
}
