package primitives

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func TestWorkerDeserializeSerialize(t *testing.T) {
	worker := Worker{
		OutPoint: OutPoint{
			TxHash: chainhash.Hash{1},
			Index:  2,
		},
		PubKey:       [48]byte{5},
		Balance:      10,
		PayeeAddress: "test2",
	}

	buf := bytes.NewBuffer([]byte{})
	err := worker.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var worker2 Worker
	if err := worker2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(worker2, worker); diff != nil {
		t.Fatal(diff)
	}
}

func TestWorkerStateDeserializeSerialize(t *testing.T) {
	workerState := WorkerState{
		Workers: map[chainhash.Hash]Worker{
			chainhash.Hash{14}: {
				OutPoint: OutPoint{
					TxHash: chainhash.Hash{1},
					Index:  2,
				},
				PubKey:       [48]byte{5},
				Balance:      10,
				PayeeAddress: "test2",
			},
		},
	}

	buf := bytes.NewBuffer([]byte{})
	err := workerState.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var workerState2 WorkerState
	if err := workerState2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(workerState2, workerState); diff != nil {
		t.Fatal(diff)
	}
}
