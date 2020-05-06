package primitives

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func TestGovernanceProposalCopy(t *testing.T) {
	gov := GovernanceProposal{
		OutPoint: OutPoint{
			TxHash: chainhash.Hash{2},
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

	gov2 := gov.Copy()

	gov.OutPoint.TxHash[0] = 1
	if gov2.OutPoint.TxHash[0] == 1 {
		t.Fatal("mutating outpoint mutates copy")
	}

	gov.GovData.Amount = 1
	if gov2.GovData.Amount == 1 {
		t.Fatal("mutating gov data mutates copy")
	}
}

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

func TestGovStateCopy(t *testing.T) {
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

	govState2 := govState.Copy()
	govState.Proposals[chainhash.Hash{14}] = GovernanceProposal{}

	if govState2.Proposals[chainhash.Hash{14}].GovData.Amount == 0 {
		t.Fatal("mutating proposals mutates copy")
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

func TestGovObjectCopy(t *testing.T) {
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
	govObject2 := govObject.Copy()

	govObject.GovID[0] = 1
	if govObject2.GovID[0] == 1 {
		t.Fatal("mutating govID mutates copy")
	}

	govObject.Amount = 2
	if govObject2.Amount == 2 {
		t.Fatal("mutating amount mutates copy")
	}

	govObject.Cycles = 3
	if govObject2.Cycles == 3 {
		t.Fatal("mutating cycles mutates copy")
	}

	govObject.PayedCycles = 4
	if govObject2.PayedCycles == 4 {
		t.Fatal("mutating payedCycles mutates copy")
	}

	govObject.BurnedUtxo.Index = 5
	if govObject2.BurnedUtxo.Index == 5 {
		t.Fatal("mutating burnedUtxo mutates copy")
	}

	govObject.Name = "asdf"
	if govObject2.Name == "asdf" {
		t.Fatal("mutating name mutates copy")
	}

	govObject.URL = "asdf2"
	if govObject2.URL == "asdf2" {
		t.Fatal("mutating url mutates copy")
	}

	govObject.PayoutAddress = "asdf3"
	if govObject2.PayoutAddress == "asdf3" {
		t.Fatal("mutating payoutaddress mutates copy")
	}

	govObject.Votes[OutPoint{
		TxHash: chainhash.Hash{9},
		Index:  10,
	}] = Vote{
		GovID:    chainhash.Hash{12},
		Approval: false,
		WorkerID: OutPoint{TxHash: chainhash.Hash{12}, Index: 13},
	}

	if govObject2.Votes[OutPoint{
		TxHash: chainhash.Hash{9},
		Index:  10,
	}].Approval == false {
		t.Fatal("mutating votes mutates copy")
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
