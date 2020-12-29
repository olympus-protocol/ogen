package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecution(t *testing.T) {
	v := testdata.FuzzExecutions(10)
	for _, e := range v {
		ser, err := e.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.Execution)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)
		assert.Equal(t, e, desc)

		assert.NoError(t, e.VerifySig())
	}

	e := new(primitives.Execution)

	assert.Equal(t, "7b45c2fff17cadc579cc85339a74b3be4525a961b07c32cd72a901f5f3d269cf", e.SignatureMessage().String())
}
