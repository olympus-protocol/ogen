package p2p_test

import (
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgProofs(t *testing.T) {
	v := new(p2p.MsgProofs)

	v.Proofs = testdata.FuzzCoinProofs(2048)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgProofs)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgProofsCmd, v.Command())
	assert.Equal(t, uint64(4761604), v.MaxPayloadLength())
}
