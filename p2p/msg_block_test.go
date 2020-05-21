package p2p

// var TestBlock = MsgBlock{
// 	Block: primitives.Block{
// 		Header:    primitives.BlockHeader{},
// 		Txs:       []primitives.Tx{TestTx, TestTx, TestTx, TestTx, TestTx, TestTx, TestTx, TestTx, TestTx, TestTx},
// 		Signature: [96]byte{},
// 	},
// }

// func TestMsgBlock_EncodeDecode(t *testing.T) {
// 	buf := bytes.NewBuffer([]byte{})
// 	err := TestBlock.Encode(buf)
// 	if err != nil {
// 		t.Errorf("unable to encode block msg")
// 	}
// 	var newBlock MsgBlock
// 	err = newBlock.Decode(buf)
// 	if err != nil {
// 		t.Errorf("unable to decode block msg")
// 	}

// 	oldBlockHash := TestBlock.Header.Hash()
// 	newBlockHash := newBlock.Header.Hash()
// 	if oldBlockHash != newBlockHash {
// 		t.Errorf("block header hashes doesn't match")
// 	}
// }
