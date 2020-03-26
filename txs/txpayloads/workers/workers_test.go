package workers_txpayload

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var mockPayloadUploadAndUpdate = PayloadUploadAndUpdate{
	Utxo:   primitives.OutPoint{},
	PubKey: [48]byte{},
	Sig:    [96]byte{},
	IP:     "1.1.1.1:8080",
}

func TestPayloadUploadAndUpdate_SerializeAndDeserialize(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadUploadAndUpdate.Serialize(buf)
	if err != nil {
		t.Errorf("TestPayloadUploadAndUpdate_SerializeAndDeserialize: %v", err.Error())
	}
	var payload PayloadUploadAndUpdate
	err = payload.Deserialize(buf)
	if err != nil {
		t.Errorf("TestPayloadUploadAndUpdate_SerializeAndDeserialize: %v", err.Error())
	}
	equal := reflect.DeepEqual(payload, mockPayloadUploadAndUpdate)
	if !equal {
		t.Errorf("TestPayloadUploadAndUpdate_SerializeAndDeserialize: should be equal = true")
	}
}

var mockPayloadRevoke = PayloadRevoke{
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	WorkerID: chainhash.Hash{},
}

func TestPayloadRevoke_SerializeAndDeserialize(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadRevoke.Serialize(buf)
	if err != nil {
		t.Errorf("TestPayloadRevoke_SerializeAndDeserialize: %v", err.Error())
	}
	var payload PayloadRevoke
	err = payload.Deserialize(buf)
	if err != nil {
		t.Errorf("TestPayloadRevoke_SerializeAndDeserialize: %v", err.Error())
	}
	equal := reflect.DeepEqual(payload, mockPayloadRevoke)
	if !equal {
		t.Errorf("TestPayloadRevoke_SerializeAndDeserialize: should be equal = true")
	}
}
