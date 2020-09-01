package p2p_test

import (
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgTxMulti(t *testing.T) {
	v := new(p2p.MsgTxMulti)
	v.Data = testdata.FuzzTxMulti(1)[0]

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgTxMulti)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgTxMultiCmd, v.Command())
	assert.Equal(t, uint64(4673), v.MaxPayloadLength())

}
