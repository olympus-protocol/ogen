package p2p_test

import (
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgExecution(t *testing.T) {
	v := new(p2p.MsgExecution)
	v = testdata.FuzzMsgExecutions(1)[0]

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgExecution)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgExecutionCmd, v.Command())
	assert.Equal(t, uint64(7352), v.MaxPayloadLength())

}
