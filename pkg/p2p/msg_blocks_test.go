package p2p_test

import (
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgBlocks(t *testing.T) {
	v := new(p2p.MsgBlocks)
	v.Blocks = testdata.FuzzBlock(32, true, true)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgBlocks)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgBlocksCmd, v.Command())
	assert.Equal(t, uint64(83886080), v.MaxPayloadLength())

}
