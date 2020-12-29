package p2p_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgProofs(t *testing.T) {
	f := fuzz.New().NilChance(0)
	v := new(p2p.MsgProofs)
	f.Fuzz(v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgProofs)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgProofsCmd, v.Command())
	assert.Equal(t, uint64(4704256), v.MaxPayloadLength())
}
