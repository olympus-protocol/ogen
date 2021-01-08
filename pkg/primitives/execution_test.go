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

		assert.LessOrEqual(t, len(ser), primitives.MaxExecutionSize)

		desc := new(primitives.Execution)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)
		assert.Equal(t, e, desc)

		assert.NoError(t, e.VerifySig())
	}

	e := new(primitives.Execution)

	assert.Equal(t, "0fee62d9e6dd5d7c3251c1d84e5e1251ae869d1535a99bdfe622b5ad79c607b1", e.SignatureMessage().String())
}
