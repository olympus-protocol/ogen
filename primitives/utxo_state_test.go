package primitives

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func TestUtxoSerializeDeserialize(t *testing.T) {
	utxo := Utxo{
		OutPoint:          OutPoint{TxHash: chainhash.Hash{1}, Index: 2},
		PrevInputsPubKeys: [][48]byte{{3}},
		Owner:             "test",
		Amount:            4,
	}

	buf := bytes.NewBuffer([]byte{})
	err := utxo.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var utxo2 Utxo
	if err := utxo2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(utxo2, utxo); diff != nil {
		t.Fatal(diff)
	}
}

func TestUtxoStateSerializeDeserialize(t *testing.T) {
	utxoState := UtxoState{
		UTXOs: map[chainhash.Hash]Utxo{
			chainhash.Hash{1}: {
				OutPoint:          OutPoint{chainhash.Hash{1}, 2},
				PrevInputsPubKeys: [][48]byte{{3}},
				Owner:             "test",
				Amount:            4,
			},
		},
	}

	buf := bytes.NewBuffer([]byte{})
	err := utxoState.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var utxoState2 UtxoState
	if err := utxoState2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(utxoState2, utxoState); diff != nil {
		t.Fatal(diff)
	}
}
