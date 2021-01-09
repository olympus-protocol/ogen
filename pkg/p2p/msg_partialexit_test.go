package p2p_test

import (
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgPartialExit(t *testing.T) {
	v := new(p2p.MsgPartialExits)
	v.Data = testdata.FuzzPartialExits(1024)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgPartialExits)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgPartialExitsCmd, v.Command())
	assert.Equal(t, uint64(200), v.MaxPayloadLength())

}
