package users_txpayload

import (
	"bytes"
	"reflect"
	"testing"
)

var mockPayloadUpload = PayloadUpload{
	PubKey: [48]byte{},
	Sig:    [96]byte{},
	Name:   "test-username",
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

var mockPayloadUpdate = PayloadUpdate{
	NewPubKey: [48]byte{},
	PubKey:    [48]byte{},
	Sig:       [96]byte{},
	Name:      "test-username",
}

func TestPayloadUpdate_SerializeAndDeserialize(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadUpdate.Serialize(buf)
	if err != nil {
		t.Errorf("TestPayloadUpdate_SerializeAndDeserialize: %v", err.Error())
	}
	var payload PayloadUpdate
	err = payload.Deserialize(buf)
	if err != nil {
		t.Errorf("TestPayloadUpdate_SerializeAndDeserialize: %v", err.Error())
	}
	equal := reflect.DeepEqual(payload, mockPayloadUpdate)
	if !equal {
		t.Errorf("TestPayloadUpdate_SerializeAndDeserialize: should be equal = true")
	}
}

var mockPayloadRevoke = PayloadRevoke{
	Sig:  [96]byte{},
	Name: "test-username",
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
