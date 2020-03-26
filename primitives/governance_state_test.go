package primitives

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func TestGovernanceProposalDeserializeSerialize(t *testing.T) {
	govProposal := GovernanceProposal{
		OutPoint: OutPoint{
			TxHash: chainhash.Hash{1},
			Index:  2,
		},
		GovData: GovObject{
			GovID:       chainhash.Hash{3},
			Amount:      4,
			Cycles:      5,
			PayedCycles: 6,
			BurnedUtxo: OutPoint{
				TxHash: chainhash.Hash{7},
				Index:  8,
			},
			Name:          "test",
			URL:           "test2",
			PayoutAddress: "test3",
			Votes: map[OutPoint]Vote{
				OutPoint{
					TxHash: chainhash.Hash{9},
					Index:  10,
				}: {
					GovID:    chainhash.Hash{11},
					Approval: true,
					WorkerID: OutPoint{TxHash: chainhash.Hash{12}, Index: 13},
				},
			},
		},
	}

	buf := bytes.NewBuffer([]byte{})
	err := govProposal.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var proposal2 GovernanceProposal
	if err := proposal2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(proposal2, govProposal); diff != nil {
		t.Fatal(diff)
	}
}

func TestGovernanceStateDeserializeSerialize(t *testing.T) {
	govState := GovernanceState{
		Proposals: map[chainhash.Hash]GovernanceProposal{
			chainhash.Hash{14}: {
				OutPoint: OutPoint{
					TxHash: chainhash.Hash{1},
					Index:  2,
				},
				GovData: GovObject{
					GovID:       chainhash.Hash{3},
					Amount:      4,
					Cycles:      5,
					PayedCycles: 6,
					BurnedUtxo: OutPoint{
						TxHash: chainhash.Hash{7},
						Index:  8,
					},
					Name:          "test",
					URL:           "test2",
					PayoutAddress: "test3",
					Votes: map[OutPoint]Vote{
						OutPoint{
							TxHash: chainhash.Hash{9},
							Index:  10,
						}: {
							GovID:    chainhash.Hash{11},
							Approval: true,
							WorkerID: OutPoint{TxHash: chainhash.Hash{12}, Index: 13},
						},
					},
				},
			},
		},
	}

	buf := bytes.NewBuffer([]byte{})
	err := govState.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var govState2 GovernanceState
	if err := govState2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(govState2, govState); diff != nil {
		t.Fatal(diff)
	}
}

func TestGovObjectSerializeDeserialize(t *testing.T) {
	govObject := GovObject{
		GovID:       chainhash.Hash{3},
		Amount:      4,
		Cycles:      5,
		PayedCycles: 6,
		BurnedUtxo: OutPoint{
			TxHash: chainhash.Hash{7},
			Index:  8,
		},
		Name:          "test",
		URL:           "test2",
		PayoutAddress: "test3",
		Votes: map[OutPoint]Vote{
			OutPoint{
				TxHash: chainhash.Hash{9},
				Index:  10,
			}: {
				GovID:    chainhash.Hash{11},
				Approval: true,
				WorkerID: OutPoint{TxHash: chainhash.Hash{12}, Index: 13},
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
