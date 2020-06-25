package serialization_test

import (
	"testing"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/test/data"
	"github.com/prysmaticlabs/go-ssz"
)

func Test_BlockNodeDiskSerialize(t *testing.T) {
	ser, err := testdata.BlockNode.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc bdb.BlockNodeDisk
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.BlockNode, desc)
	if !equal {
		t.Fatal("error: serialize BlockNodeDisk")
	}
}
