package p2p_test

import (
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgDeposits(t *testing.T) {
	v := new(p2p.MsgDeposits)
	v.Data = testdata.FuzzDeposit(1024, true)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgDeposits)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgDepositsCmd, v.Command())
	assert.Equal(t, uint64(308*1024), v.MaxPayloadLength())

}
