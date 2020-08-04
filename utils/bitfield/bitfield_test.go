package bitfcheck_test

import (
	"github.com/olympus-protocol/ogen/utils/bitfield"
	prysmbf "github.com/prysmaticlabs/go-bitfield"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_PrysmBitfield(t *testing.T) {
	bf := prysmbf.NewBitlist(4 * 8)

	bitfcheck.Set(bf, 32)

	assert.True(t, bitfcheck.Get(bf, 32))
}
