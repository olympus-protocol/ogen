package primitives_test

import (
	"testing"

	"github.com/olympus-protocol/ogen/primitives"
	"github.com/prysmaticlabs/go-ssz"
)

var validator = primitives.Validator{
	Balance:          100000,
	PubKey:           []byte{0x50, 0x60},
	PayeeAddress:     [20]byte{0x99},
	Status:           primitives.StatusExitedWithoutPenalty,
	FirstActiveEpoch: 100,
	LastActiveEpoch:  500,
}

func Test_ValidatorsSerialize(t *testing.T) {
	ser, err := validator.Marshal()
	if err != nil {
		t.Error(err)
	}
	var desc primitives.Validator
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Error(err)
	}
	equal := ssz.DeepEqual(validator, desc)
	if !equal {
		t.Error("marshal/unmarshal failed for Validator")
	}
}
