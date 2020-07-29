package p2p_test

import (
	"bytes"
	"testing"

	"github.com/olympus-protocol/ogen/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
)

func Test_MessageHeaderSerialize(t *testing.T) {

	ser, err := testdata.Header.Marshal()

	assert.NoError(t, err)

	var desc p2p.MessageHeader

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, testdata.Header, desc)
}

func Test_MsgGetAddrSerialize(t *testing.T) {

	ser, err := testdata.MsgGetAddr.Marshal()

	assert.NoError(t, err)

	var desc p2p.MsgGetAddr

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, testdata.MsgGetAddr, desc)

}

func Test_MsgAddrSerialize(t *testing.T) {

	ser, err := testdata.MsgAddr.Marshal()

	assert.NoError(t, err)

	var desc p2p.MsgAddr

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, testdata.MsgAddr, desc)

}

func Test_MsgGetBlocksSerialize(t *testing.T) {

	ser, err := testdata.MsgGetBlocks.Marshal()

	assert.NoError(t, err)

	var desc p2p.MsgGetBlocks

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, testdata.MsgGetBlocks, desc)
}

func Test_MsgVersionSerialize(t *testing.T) {

	ser, err := testdata.MsgVersion.Marshal()

	assert.NoError(t, err)

	var desc p2p.MsgVersion

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, testdata.MsgVersion, desc)
}

func Test_MsgBlocksSerialize(t *testing.T) {

	ser, err := testdata.MsgBlocks.Marshal()

	assert.NoError(t, err)

	var desc p2p.MsgBlocks

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, testdata.MsgBlocks, desc)

	// Convert the block pointers to a slice of blocks
	//var expectedBlocks, serializedBlocks []primitives.Block
	//for _, b := range testdata.MsgBlocks.Blocks {
	//	expectedBlocks = append(expectedBlocks, *b)
	//}
	//for _, b := range desc.Blocks {
	//	serializedBlocks = append(serializedBlocks, *b)
	//}
	//assert.Equal(t, serializedBlocks, expectedBlocks)
}

func Test_MsgWithHeaderSerialize(t *testing.T) {

	buf := bytes.NewBuffer([]byte{})

	err := p2p.WriteMessage(buf, &testdata.MsgAddr, 333)

	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 333)

	assert.NoError(t, err)

	assert.Equal(t, msg, &testdata.MsgAddr)
}
