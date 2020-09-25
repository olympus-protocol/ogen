package p2p_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgAddr(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(32, 32)
	v := new(p2p.MsgAddr)
	f.Fuzz(v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgAddr)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgAddrCmd, v.Command())
	assert.Equal(t, uint64(16388), v.MaxPayloadLength())

}
