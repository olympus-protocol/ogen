package coins_txpayload

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var mockInputCoinBase = Input{
	PrevOutpoint: primitives.OutPoint{},
	Sig:          [96]byte{},
	PubKey:       [48]byte{},
}

var mockInput = Input{
	PrevOutpoint: primitives.OutPoint{TxHash: chainhash.Hash{}, Index: 10},
	Sig:          [96]byte{1, 1, 1},
	PubKey:       [48]byte{1, 1, 1},
}

func TestInput_SerializeAndDeserialize(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockInputCoinBase.Serialize(buf)
	if err != nil {
		t.Errorf("TestInput_SerializeAndDeserialize: %v", err.Error())
	}
	var deserializedInput Input
	err = deserializedInput.Deserialize(buf)
	if err != nil {
		t.Errorf("TestInput_SerializeAndDeserialize: %v", err.Error())
	}
	equal := reflect.DeepEqual(deserializedInput, mockInputCoinBase)
	if !equal {
		t.Errorf("TestInput_SerializeAndDeserialize: should be equal = true")
	}
}

func TestNewInput(t *testing.T) {
	input := NewInput(mockInputCoinBase.PrevOutpoint, mockInputCoinBase.Sig, mockInputCoinBase.PubKey)
	equal := reflect.DeepEqual(input, mockInputCoinBase)
	if !equal {
		t.Errorf("TestNewInput: should be equal = true")
	}
}

var mockOutput = Output{
	Value:   0,
	Address: "mock-address",
}

func TestOutput_SerializeAndDeserialize(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockOutput.Serialize(buf)
	if err != nil {
		t.Errorf("TestOutput_SerializeAndDeserialize: %v", err.Error())
	}
	var deserializedOutput Output
	err = deserializedOutput.Deserialize(buf)
	if err != nil {
		t.Errorf("TestOutput_SerializeAndDeserialize: %v", err.Error())
	}
	equal := reflect.DeepEqual(deserializedOutput, mockOutput)
	if !equal {
		t.Errorf("TestOutput_SerializeAndDeserialize: should be equal = true")
	}
}

func TestNewOutput(t *testing.T) {
	output := NewOutput(mockOutput.Value, mockOutput.Address)
	equal := reflect.DeepEqual(output, mockOutput)
	if !equal {
		t.Errorf("TestNewInput: should be equal = true")
	}
}

var mockPayloadTransfer = PayloadTransfer{
	AggSig: [96]byte{},
	TxIn:   []Input{mockInputCoinBase, mockInput},
	TxOut:  []Output{mockOutput},
}

func TestPayloadTransfer_SerializeAndDeserialize(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadTransfer.Serialize(buf)
	if err != nil {
		t.Errorf("TestPayloadTransfer_SerializeAndDeserialize: %v", err.Error())
	}
	var deserializedPayloadTransfer PayloadTransfer
	err = deserializedPayloadTransfer.Deserialize(buf)
	if err != nil {
		t.Errorf("TestPayloadTransfer_SerializeAndDeserialize: %v", err.Error())
	}
	equal := reflect.DeepEqual(deserializedPayloadTransfer, mockPayloadTransfer)
	if !equal {
		t.Errorf("TestPayloadTransfer_SerializeAndDeserialize: should be equal = true")
	}
}
