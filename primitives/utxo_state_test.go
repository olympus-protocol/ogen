package primitives

// func TestUtxoCopy(t *testing.T) {
// 	utxo := Utxo{
// 		OutPoint:          OutPoint{TxHash: chainhash.Hash{0}, Index: 0},
// 		PrevInputsPubKeys: [][48]byte{{0}},
// 		Owner:             "test",
// 		Amount:            0,
// 	}

// 	u2 := utxo.Copy()

// 	utxo.Amount = 1
// 	if u2.Amount == 1 {
// 		t.Fatal("mutating amount mutates copy")
// 	}

// 	utxo.PrevInputsPubKeys[0][0] = 1
// 	if u2.PrevInputsPubKeys[0][0] == 1 {
// 		t.Fatal("mutating pubkeys mutates copy")
// 	}

// 	utxo.Owner = "test2"
// 	if u2.Owner == "test2" {
// 		t.Fatal("mutating owner mutates copy")
// 	}

// 	utxo.OutPoint.TxHash[0] = 1
// 	if u2.OutPoint.TxHash[0] == 1 {
// 		t.Fatal("mutating outpoint mutates copy")
// 	}
// }

// func TestUtxoSerializeDeserialize(t *testing.T) {
// 	utxo := Utxo{
// 		OutPoint:          OutPoint{TxHash: chainhash.Hash{1}, Index: 2},
// 		PrevInputsPubKeys: [][48]byte{{3}},
// 		Owner:             "test",
// 		Amount:            4,
// 	}

// 	buf := bytes.NewBuffer([]byte{})
// 	err := utxo.Serialize(buf)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	var utxo2 Utxo
// 	if err := utxo2.Deserialize(buf); err != nil {
// 		t.Fatal(err)
// 	}

// 	if diff := deep.Equal(utxo2, utxo); diff != nil {
// 		t.Fatal(diff)
// 	}
// }

// func TestUtxoStateCopy(t *testing.T) {
// 	utxoState := UtxoState{
// 		UTXOs: map[chainhash.Hash]Utxo{
// 			chainhash.Hash{1}: {
// 				OutPoint:          OutPoint{chainhash.Hash{1}, 2},
// 				PrevInputsPubKeys: [][48]byte{{3}},
// 				Owner:             "test",
// 				Amount:            4,
// 			},
// 		},
// 	}

// 	s2 := utxoState.Copy()

// 	utxoState.UTXOs[chainhash.Hash{1}] = Utxo{
// 		Amount: 1,
// 	}
// 	if s2.UTXOs[chainhash.Hash{1}].Amount == 1 {
// 		t.Fatal("mutating UTXOs mutates copy")
// 	}
// }

// func TestUtxoStateSerializeDeserialize(t *testing.T) {
// 	utxoState := UtxoState{
// 		UTXOs: map[chainhash.Hash]Utxo{
// 			chainhash.Hash{1}: {
// 				OutPoint:          OutPoint{chainhash.Hash{1}, 2},
// 				PrevInputsPubKeys: [][48]byte{{3}},
// 				Owner:             "test",
// 				Amount:            4,
// 			},
// 		},
// 	}

// 	buf := bytes.NewBuffer([]byte{})
// 	err := utxoState.Serialize(buf)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	var utxoState2 UtxoState
// 	if err := utxoState2.Deserialize(buf); err != nil {
// 		t.Fatal(err)
// 	}

// 	if diff := deep.Equal(utxoState2, utxoState); diff != nil {
// 		t.Fatal(diff)
// 	}
// }
