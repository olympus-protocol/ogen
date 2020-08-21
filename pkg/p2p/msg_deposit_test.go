package p2p_test

import (
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgDeposit(t *testing.T) {
	v := new(p2p.MsgDeposit)
	v.Data = testdata.FuzzDeposit(1, true)[0]

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgDeposit)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgDepositCmd, v.Command())
	assert.Equal(t, uint64(308), v.MaxPayloadLength())

}
