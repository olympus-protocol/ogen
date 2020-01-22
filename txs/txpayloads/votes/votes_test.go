package votes_txpayload

import (
	"bytes"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"reflect"
	"testing"
)

var mockPayloadUploadAndUpdate = PayloadUploadAndUpdate{
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
	Approval: false,
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
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
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
