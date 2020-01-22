package gov_txpayload

import (
	"bytes"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"reflect"
	"testing"
)

var mockPayloadUpload = PayloadUpload{
	BurnedUtxo:    p2p.OutPoint{},
	PubKey:        [48]byte{},
	Sig:           [96]byte{},
	Name:          "mock-name",
	URL:           "https://test.name",
	PayoutAddress: "TestAddr",
	Amount:        10000,
	Cycles:        10,
}

func TestPayloadUpload_SerializeAndDeserialize(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadUpload.Serialize(buf)
	if err != nil {
		t.Errorf("TestPayloadUpload_SerializeAndDeserialize: %v", err.Error())
	}
	var payload PayloadUpload
	err = payload.Deserialize(buf)
	if err != nil {
		t.Errorf("TestPayloadUpload_SerializeAndDeserialize: %v", err.Error())
	}
	equal := reflect.DeepEqual(payload, mockPayloadUpload)
	if !equal {
		t.Errorf("TestPayloadUpload_SerializeAndDeserialize: should be equal = true")
	}
}

var mockPayloadRevoke = PayloadRevoke{
	GovID:  chainhash.Hash{},
	PubKey: [48]byte{},
	Sig:    [96]byte{},
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
