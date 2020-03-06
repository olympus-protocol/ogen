package gov

import (
	"bytes"
	"github.com/go-test/deep"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"testing"
)

func TestGovObjectSerializeDeserialize(t *testing.T) {
	govObject := GovObject{
		GovID:       chainhash.Hash{3},
		Amount:      4,
		Cycles:      5,
		PayedCycles: 6,
		BurnedUtxo: p2p.OutPoint{
			TxHash: chainhash.Hash{7},
			Index:  8,
		},
		Name:          "test",
		URL:           "test2",
		PayoutAddress: "test3",
		Votes: map[p2p.OutPoint]Vote{
			p2p.OutPoint{
				TxHash: chainhash.Hash{9},
				Index:  10,
			}: {
				GovID:    chainhash.Hash{11},
				Approval: true,
				WorkerID: p2p.OutPoint{TxHash: chainhash.Hash{12}, Index: 13},
			},
		},
	}

	buf := bytes.NewBuffer([]byte{})
	err := govObject.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var govObject2 GovObject
	if err := govObject2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(govObject2, govObject); diff != nil {
		t.Fatal(diff)
	}
}