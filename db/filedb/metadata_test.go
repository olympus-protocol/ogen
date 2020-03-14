package filedb

import (
	"bytes"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"testing"
)

var metaData = MetaData{
	Version:     1001001,
	Timestamp:   1571761181,
	Name:        "test-meta-data-big-name",
	MaxElemSize: 200,
}

func TestMetaData_Serialize(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := metaData.serialize(buf)
	if err != nil {
		t.Errorf("Error %v", err.Error())
	}
	if metaData.serializedSize() != uint64(len(buf.Bytes())) {
		t.Errorf("MetaData Encode error: serialized size doesn't match")
	}
	copyBuf := bytes.NewBuffer(buf.Bytes())
	metaDataSizeSerialized, err := serializer.ReadVarInt(copyBuf)
	if err != nil {
		t.Errorf("MetaData Encode error: unable to decode serialized size")
	}
	if metaData.serializedSize() != metaDataSizeSerialized {
		t.Errorf("MetaData Encode error: serialized size doesn't match")
	}
	if len(buf.Bytes()) == 0 {
		t.Errorf("MetaData Encode error: Empty bytes from metadata")
	}
	var newMetaData MetaData
	err = newMetaData.deserialize(buf)
	if err != nil {
		t.Errorf("MetaData Decode error: %v", err.Error())
	}
	if newMetaData.Name != metaData.Name {
		t.Errorf("MetaData Decode error: name doesn't match")
	}
	if newMetaData.Timestamp != metaData.Timestamp {
		t.Errorf("MetaData Decode error: timestamp doesn't match")
	}
	if newMetaData.Version != metaData.Version {
		t.Errorf("MetaData Decode error: version doesn't match")
	}
	if newMetaData.MaxElemSize != metaData.MaxElemSize {
		t.Errorf("MetaData Decode error: max element size doesn't match")
	}
}
