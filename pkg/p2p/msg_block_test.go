package p2p_test

import (
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgBlock(t *testing.T) {
	v := new(p2p.MsgBlock)
	v.Data = testdata.FuzzBlock(1, true, true)[0]

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgBlock)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgBlockCmd, v.Command())
	assert.Equal(t, uint64(2278442), v.MaxPayloadLength())

}
