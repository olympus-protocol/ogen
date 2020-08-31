package p2p_test

import (
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgExits(t *testing.T) {
	v := new(p2p.MsgExits)
	v.Data = testdata.FuzzExits(1024)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgExits)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgExitsCmd, v.Command())
	assert.Equal(t, uint64(196608), v.MaxPayloadLength())

}
