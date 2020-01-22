package workers_txpayload

import (
	"bytes"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/utils/chainhash"
	"reflect"
	"testing"
)

var mockPayloadUploadAndUpdate = PayloadUploadAndUpdate{
	Utxo:   p2p.OutPoint{},
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
