package p2p

// import (
// 	"bytes"
// 	"testing"
// 	"time"

// 	"github.com/olympus-protocol/ogen/primitives"
// )

// var (
// 	TestTx = primitives.Tx{
// 		TxVersion: 1,
// 		TxType:    primitives.TxCoins,
// 		TxAction:  primitives.Transfer,
// 		Time:      time.Unix(1572830409, 0).Unix(),
// 	}
// )

// func TestMsgTx_EncodeDecode(t *testing.T) {
// 	buf := bytes.NewBuffer([]byte{})
// 	err := TestTx.Encode(buf)
// 	if err != nil {
// 		t.Errorf("unable to encode tx msg")
// 	}
// 	var newTx primitives.Tx
// 	err = newTx.Decode(buf)
// 	if err != nil {
// 		t.Errorf("unable to decode tx msg")
// 	}
// 	testTxHash := TestTx.Hash()
// 	newTxHash := newTx.Hash()

// 	if !testTxHash.IsEqual(&newTxHash) {
// 		t.Fatal("expected old hash to match new hash")
// 	}
// }
