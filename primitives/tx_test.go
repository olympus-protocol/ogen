package primitives_test

import (
	"errors"
	"testing"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/prysmaticlabs/go-ssz"
)

var multiSig bls.Multisig
var randPub []byte

func init() {
	randPub = bls.RandKey().PublicKey().Marshal()
	multiPub := bls.NewMultipub([]*bls.PublicKey{bls.RandKey().PublicKey()}, 1)
	multiSig = *bls.NewMultisig(*multiPub)
}

var payloadSingle = &primitives.TransferSinglePayload{
	To:            [20]byte{0x0, 0x2, 0x5},
	FromPublicKey: randPub,
	Amount:        10000,
	Nonce:         9999,
	Fee:           1,
	Signature:     bls.NewAggregateSignature().Marshal(),
}

var payloadMulti = &primitives.TransferMultiPayload{
	To:        [20]byte{0x0, 0x2, 0x5},
	Amount:    10000,
	Nonce:     9999,
	Fee:       1,
	Signature: multiSig,
}

var txMulti = primitives.Tx{
	Version: 1,
	Type:    primitives.TxTransferMulti,
	Payload: payloadMulti,
}

var txSingle = primitives.Tx{
	Version: 1,
	Type:    primitives.TxTransferSingle,
	Payload: payloadSingle,
}

func Test_TxSerialization(t *testing.T) {
	err := serSinglePayload()
	if err != nil {
		t.Error(err)
	}
	err = serMultiPayload()
	if err != nil {
		t.Error(err)
	}
	err = serMultiTx()
	if err != nil {
		t.Error(err)
	}
	err = serSingleTx()
	if err != nil {
		t.Error(err)
	}
}

func serSinglePayload() error {
	ser, err := payloadSingle.Marshal()
	if err != nil {
		return err
	}
	var desc primitives.TransferSinglePayload
	err = desc.Unmarshal(ser)
	if err != nil {
		return err
	}
	equal := ssz.DeepEqual(desc, payloadSingle)
	if !equal {
		return errors.New("marshal/unmarshal failed for TransferSinglePayload")
	}
	return nil
}

func serMultiPayload() error {
	ser, err := payloadMulti.Marshal()
	if err != nil {
		return err
	}
	var desc primitives.TransferMultiPayload
	err = desc.Unmarshal(ser)
	if err != nil {
		return err
	}
	equal := ssz.DeepEqual(desc, payloadMulti)
	if !equal {
		return errors.New("marshal/unmarshal failed for TransferMultiPayload")
	}
	return nil
}

func serSingleTx() error {
	ser, err := txSingle.Marshal()
	if err != nil {
		return err
	}
	var desc primitives.Tx
	err = desc.Unmarshal(ser)
	if err != nil {
		return err
	}
	equal := ssz.DeepEqual(desc, txSingle)
	if !equal {
		return errors.New("marshal/unmarshal failed for Tx single")
	}
	return nil
}

func serMultiTx() error {
	ser, err := txMulti.Marshal()
	if err != nil {
		return err
	}
	var desc primitives.Tx
	err = desc.Unmarshal(ser)
	if err != nil {
		return err
	}
	equal := ssz.DeepEqual(desc, txMulti)
	if !equal {
		return errors.New("marshal/unmarshal failed for Tx multi")
	}
	return nil
}
