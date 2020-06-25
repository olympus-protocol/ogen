package serialization_test

import (
	"testing"

	"github.com/olympus-protocol/ogen/bls"
	testdata "github.com/olympus-protocol/ogen/test/data"
	"github.com/prysmaticlabs/go-ssz"
)

func Test_CombinedSignatureSerialize(t *testing.T) {
	ser, err := testdata.CombinedSignature.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc bls.CombinedSignature
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.CombinedSignature, desc)
	if !equal {
		t.Fatal("error: serialize CombinedSignature")
	}
}

func Test_MultipubSerialize(t *testing.T) {
	ser := testdata.Multipub.Marshal()
	var desc bls.Multipub
	err := desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.Multipub, desc)
	if !equal {
		t.Fatal("error: serialize Multipub")
	}
}

func Test_MultisigSerialize(t *testing.T) {
	ser, err := testdata.Multisig.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc bls.Multisig
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.Multisig, desc)
	if !equal {
		t.Fatal("error: serialize Multisig")
	}
}
