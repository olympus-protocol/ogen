package p2p_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgGetBlocks(t *testing.T) {
	f := fuzz.New().NilChance(0)
	v := new(p2p.MsgGetBlocks)
	f.Fuzz(v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MsgGetBlocks)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)

	assert.Equal(t, p2p.MsgGetBlocksCmd, v.Command())
	assert.Equal(t, uint64(32), v.MaxPayloadLength())

}
