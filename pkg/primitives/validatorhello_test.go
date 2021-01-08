package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidatorHello(t *testing.T) {
	d := testdata.FuzzValidatorHello(10)

	for _, c := range d {

		ser, err := c.Marshal()
		assert.NoError(t, err)

		assert.LessOrEqual(t, len(ser), primitives.MaxValidatorHelloMessageSize)

		desc := new(primitives.ValidatorHelloMessage)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)
	}

}
