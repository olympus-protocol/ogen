package serialization_test

import (
	"bytes"
	"testing"

	"github.com/olympus-protocol/ogen/p2p"
	testdata "github.com/olympus-protocol/ogen/test/data"
	"github.com/prysmaticlabs/go-ssz"
)

func Test_MessageHeaderSerialize(t *testing.T) {
	ser, err := testdata.Header.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc p2p.MessageHeader
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.Header, desc)
	if !equal {
		t.Fatal("error: serialize MessageHeader")
	}
}

func Test_MsgGetAddrSerialize(t *testing.T) {
	ser, err := testdata.MsgGetAddr.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc p2p.MsgGetAddr
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.MsgGetAddr, desc)
	if !equal {
		t.Fatal("error: serialize MsgAddr")
	}
}

func Test_MsgAddrSerialize(t *testing.T) {
	ser, err := testdata.MsgAddr.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc p2p.MsgAddr
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.MsgAddr, desc)
	if !equal {
		t.Fatal("error: serialize MsgAddr")
	}
}

func Test_MsgGetBlocksSerialize(t *testing.T) {
	ser, err := testdata.MsgGetBlocks.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc p2p.MsgGetBlocks
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.MsgGetBlocks, desc)
	if !equal {
		t.Fatal("error: serialize MsgGetBlocks")
	}
}

func Test_MsgVersionSerialize(t *testing.T) {
	ser, err := testdata.MsgVersion.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc p2p.MsgVersion
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.MsgVersion, desc)
	if !equal {
		t.Fatal("error: serialize MsgVersion")
	}
}

func Test_MsgBlocksSerialize(t *testing.T) {
	ser, err := testdata.MsgBlocks.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc p2p.MsgBlocks
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.MsgBlocks, desc)
	if !equal {
		t.Fatal("error: serialize MsgBlocks")
	}
}

func Test_MsgWithHeaderSerialize(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := p2p.WriteMessage(buf, &testdata.MsgBlocks, 333)
	if err != nil {
		t.Error(err)
	}
	msg, err := p2p.ReadMessage(buf, 333)
	if err != nil {
		t.Error(err)
	}
	equal := ssz.DeepEqual(msg.(*p2p.MsgBlocks), &testdata.MsgBlocks)
	if !equal {
		t.Error("error: serialize MsgWithHeader")
	}
}
