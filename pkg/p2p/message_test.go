package p2p_test

import (
	"bytes"
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessageHeader(t *testing.T) {
	f := fuzz.New().NilChance(0)
	d := new(p2p.MessageHeader)
	f.Fuzz(d)

	ser, err := d.Marshal()
	assert.NoError(t, err)

	desc := new(p2p.MessageHeader)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, d, desc)
}

func TestReadWriteMessage(t *testing.T) {
	v := new(p2p.MsgBlock)

	v.Data = testdata.FuzzBlock(1, true, true)[0]

	wrongbuf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(wrongbuf, v, 333)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(wrongbuf, 222)
	assert.Nil(t, msg)
	assert.Equal(t, p2p.ErrorNetMismatch, err)

	buf := bytes.NewBuffer([]byte{})

	err = p2p.WriteMessage(buf, v, 333)
	assert.NoError(t, err)

	msg, err = p2p.ReadMessage(buf, 333)
	assert.NoError(t, err)

	blockMsg, ok := msg.(*p2p.MsgBlock)
	assert.True(t, ok)
	assert.Equal(t, v, blockMsg)
}

func TestMsgTypeCreation(t *testing.T) {
	createMsgVersion(t)
	createMsgGetBlocks(t)
	createMsgTx(t)
	createMsgTxMulti(t)
	createMsgBlock(t)
	createMsgDeposit(t)
	createMsgDeposits(t)
	createMsgExit(t)
	createMsgExits(t)
	createMsgVote(t)
	createMsgValidatorStart(t)
	createMsgGovernance(t)
	createMsgFinalization(t)
}

func createMsgVersion(t *testing.T) {
	v := new(p2p.MsgVersion)
	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgVersion)
	assert.True(t, ok)
}

func createMsgGetBlocks(t *testing.T) {
	v := new(p2p.MsgGetBlocks)
	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgGetBlocks)
	assert.True(t, ok)
}

func createMsgBlock(t *testing.T) {
	v := new(p2p.MsgBlock)
	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgBlock)
	assert.True(t, ok)
}

func createMsgDeposit(t *testing.T) {
	v := new(p2p.MsgDeposit)
	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgDeposit)
	assert.True(t, ok)
}

func createMsgDeposits(t *testing.T) {
	v := new(p2p.MsgDeposits)
	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgDeposits)
	assert.True(t, ok)
}

func createMsgExit(t *testing.T) {
	v := new(p2p.MsgExit)
	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgExit)
	assert.True(t, ok)
}

func createMsgExits(t *testing.T) {
	v := new(p2p.MsgExits)
	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgExits)
	assert.True(t, ok)
}

func createMsgVote(t *testing.T) {
	v := new(p2p.MsgVote)
	buf := bytes.NewBuffer([]byte{})
	v.Data = testdata.FuzzMultiValidatorVote(1, true, true)[0]
	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgVote)
	assert.True(t, ok)
}

func createMsgTx(t *testing.T) {
	v := new(p2p.MsgTx)
	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgTx)
	assert.True(t, ok)
}

func createMsgTxMulti(t *testing.T) {
	v := new(p2p.MsgMultiSignatureTx)
	buf := bytes.NewBuffer([]byte{})
	v.Data = testdata.FuzzTxMulti(1)[0]
	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgMultiSignatureTx)
	assert.True(t, ok)
}

func createMsgValidatorStart(t *testing.T) {
	v := new(p2p.MsgValidatorStart)
	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgValidatorStart)
	assert.True(t, ok)
}

func createMsgGovernance(t *testing.T) {
	v := new(p2p.MsgGovernance)
	buf := bytes.NewBuffer([]byte{})
	v.Data = testdata.FuzzGovernanceVote(1)[0]
	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgGovernance)
	assert.True(t, ok)
}

func createMsgFinalization(t *testing.T) {
	v := new(p2p.MsgFinalization)
	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, v, 1)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 1)
	assert.NoError(t, err)

	_, ok := msg.(*p2p.MsgFinalization)
	assert.True(t, ok)
}
