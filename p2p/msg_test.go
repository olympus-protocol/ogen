package p2p_test

import (
	"bytes"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/stretchr/testify/assert"
)

func Test_MessageHeaderSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(p2p.MessageHeader)
	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(p2p.MessageHeader)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgGetAddrSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(p2p.MsgGetAddr)
	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(p2p.MsgGetAddr)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)

}

func Test_MsgAddrSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(p2p.MsgAddr)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(p2p.MsgGetAddr)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)

}

func Test_MsgGetBlocksSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)

	v := new(p2p.MsgGetBlocks)
	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(p2p.MsgGetAddr)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgVersionSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)

	v := new(p2p.MsgVersion)
	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(p2p.MsgVersion)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgBlocksSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)

	v := new(p2p.MsgBlocks)
	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(p2p.MsgBlocks)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgWithHeaderSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)

	v := new(p2p.MsgAddr)
	f.Fuzz(v)

	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, v, 333)

	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 333)

	assert.NoError(t, err)

	assert.Equal(t, msg, v)
}
