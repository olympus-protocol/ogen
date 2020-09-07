package p2p_test

import (
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgSyncEnd(t *testing.T) {
	v := new(p2p.MsgSyncEnd)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgSyncEnd)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgSyncEndCmd, v.Command())
	assert.Equal(t, uint64(0), v.MaxPayloadLength())

}
