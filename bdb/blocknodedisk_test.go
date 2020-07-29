package bdb_test

import (
	"reflect"
	"testing"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/test"
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
	equal := reflect.DeepEqual(testdata.BlockNode, desc)
	if !equal {
		t.Fatal("error: serialize BlockNodeDisk")
	}
}
