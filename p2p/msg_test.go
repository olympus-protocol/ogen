package p2p_test

import (
	"bytes"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/p2p"
)

func Test_MessageHeaderSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v p2p.MessageHeader
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc p2p.MessageHeader
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgGetAddrSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v p2p.MsgGetAddr
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc p2p.MsgGetAddr
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgAddrSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v p2p.MsgAddr
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc p2p.MsgAddr
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgGetBlocksSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v p2p.MsgGetBlocks
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc p2p.MsgGetBlocks
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgVersionSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v p2p.MsgVersion
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc p2p.MsgVersion
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgBlocksSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v p2p.MsgBlocks
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc p2p.MsgBlocks
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgWithHeaderSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v p2p.MsgBlocks
	f.Fuzz(&v)

	buf := bytes.NewBuffer([]byte{})
	err := p2p.WriteMessage(buf, &v, 333)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 333)
	assert.NoError(t, err)

	assert.Equal(t, msg.(*p2p.MsgBlocks), &v)
}
