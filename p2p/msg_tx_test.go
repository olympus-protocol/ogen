package p2p

import (
	"bytes"
	"testing"
	"time"
)

var (
	TestTx = MsgTx{
		TxVersion: 1,
		TxType:    Coins,
		TxAction:  Transfer,
		Time:      time.Unix(1572830409, 0).Unix(),
	}
)

func TestMsgTx_EncodeDecode(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := TestTx.Encode(buf)
	if err != nil {
		t.Errorf("unable to encode tx msg")
	}
	var newTx MsgTx
	err = newTx.Decode(buf)
	if err != nil {
		t.Errorf("unable to decode tx msg")
	}
	testTxHash, err := TestTx.TxHash()
	if err != nil {
		t.Errorf("unable to calculate old tx hash")
	}
	newTxHash, err := newTx.TxHash()
	if err != nil {
		t.Errorf("unable to calculate new tx hash")
	}
	if testTxHash != newTxHash {
		t.Errorf("tx hash doesn't match")
	}
}
