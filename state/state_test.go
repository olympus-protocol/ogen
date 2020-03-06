package state

import (
	"bytes"
	"github.com/go-test/deep"
	"github.com/olympus-protocol/ogen/gov"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/users"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/workers"
	"testing"
)

func TestGovernanceProposalDeserializeSerialize(t *testing.T) {
	govProposal := GovernanceProposal{
		OutPoint: p2p.OutPoint{
			TxHash: chainhash.Hash{1},
			Index:  2,
		},
		GovData: gov.GovObject{
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
			Votes: map[p2p.OutPoint]gov.Vote{
				p2p.OutPoint{
					TxHash: chainhash.Hash{9},
					Index:  10,
				}: {
					GovID:    chainhash.Hash{11},
					Approval: true,
					WorkerID: p2p.OutPoint{TxHash: chainhash.Hash{12}, Index: 13},
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
				OutPoint: p2p.OutPoint{
					TxHash: chainhash.Hash{1},
					Index:  2,
				},
				GovData: gov.GovObject{
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
					Votes: map[p2p.OutPoint]gov.Vote{
						p2p.OutPoint{
							TxHash: chainhash.Hash{9},
							Index:  10,
						}: {
							GovID:    chainhash.Hash{11},
							Approval: true,
							WorkerID: p2p.OutPoint{TxHash: chainhash.Hash{12}, Index: 13},
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


func TestUtxoSerializeDeserialize(t *testing.T) {
	utxo := Utxo{
		OutPoint:          p2p.OutPoint{TxHash: chainhash.Hash{1}, Index: 2},
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
				OutPoint:          p2p.OutPoint{chainhash.Hash{1}, 2},
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

func TestWorkerDeserializeSerialize(t *testing.T) {
	worker := Worker{
		OutPoint: p2p.OutPoint{
			TxHash: chainhash.Hash{1},
			Index:  2,
		},
		WorkerData: workers.Worker{
			WorkerID:          p2p.OutPoint{
				TxHash: chainhash.Hash{3},
				Index:  4,
			},
			PubKey:            [48]byte{5},
			LastBlockAssigned: 6,
			NextBlockAssigned: 7,
			Score:             8,
			Version:           9,
			Protocol:          10,
			IP:                "test",
			PayeeAddress:      "test2",
		},
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
				OutPoint: p2p.OutPoint{
					TxHash: chainhash.Hash{1},
					Index:  2,
				},
				WorkerData: workers.Worker{
					WorkerID:          p2p.OutPoint{
						TxHash: chainhash.Hash{3},
						Index:  4,
					},
					PubKey:            [48]byte{5},
					LastBlockAssigned: 6,
					NextBlockAssigned: 7,
					Score:             8,
					Version:           9,
					Protocol:          10,
					IP:                "test",
					PayeeAddress:      "test2",
				},
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

func TestUserDeserializeSerialize(t *testing.T) {
	user := User{
		OutPoint: p2p.OutPoint{
			TxHash: chainhash.Hash{1},
			Index:  2,
		},
		UserData: users.User{
			PubKey: [48]byte{3},
			Name:   "4",
		},
	}

	buf := bytes.NewBuffer([]byte{})
	err := user.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var user2 User
	if err := user2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(user2, user); diff != nil {
		t.Fatal(diff)
	}
}


func TestUserStateDeserializeSerialize(t *testing.T) {
	userState := UserState{
		Users: map[chainhash.Hash]User{
			chainhash.Hash{14}: {
				OutPoint: p2p.OutPoint{
					TxHash: chainhash.Hash{1},
					Index:  2,
				},
				UserData: users.User{
					PubKey: [48]byte{3},
					Name:   "4",
				},
			},
		},
	}

	buf := bytes.NewBuffer([]byte{})
	err := userState.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var userState2 UserState
	if err := userState2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(userState2, userState); diff != nil {
		t.Fatal(diff)
	}
}

func TestStateDeserializeSerialize(t *testing.T) {
	state := State{
		UtxoState: UtxoState{
			UTXOs: map[chainhash.Hash]Utxo{
				chainhash.Hash{1}: {
					OutPoint:          p2p.OutPoint{chainhash.Hash{1}, 2},
					PrevInputsPubKeys: [][48]byte{{3}},
					Owner:             "test",
					Amount:            4,
				},
			},
		},
		GovernanceState: GovernanceState{
			Proposals: map[chainhash.Hash]GovernanceProposal{
				chainhash.Hash{14}: {
					OutPoint: p2p.OutPoint{
						TxHash: chainhash.Hash{1},
						Index:  2,
					},
					GovData: gov.GovObject{
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
						Votes: map[p2p.OutPoint]gov.Vote{
							p2p.OutPoint{
								TxHash: chainhash.Hash{9},
								Index:  10,
							}: {
								GovID:    chainhash.Hash{11},
								Approval: true,
								WorkerID: p2p.OutPoint{TxHash: chainhash.Hash{12}, Index: 13},
							},
						},
					},
				},
			},
		},
		UserState: UserState{
			Users: map[chainhash.Hash]User{
				chainhash.Hash{14}: {
					OutPoint: p2p.OutPoint{
						TxHash: chainhash.Hash{1},
						Index:  2,
					},
					UserData: users.User{
						PubKey: [48]byte{3},
						Name:   "4",
					},
				},
			},
		},
		WorkerState: WorkerState{
			Workers: map[chainhash.Hash]Worker{
				chainhash.Hash{14}: {
					OutPoint: p2p.OutPoint{
						TxHash: chainhash.Hash{1},
						Index:  2,
					},
					WorkerData: workers.Worker{
						WorkerID: p2p.OutPoint{
							TxHash: chainhash.Hash{3},
							Index:  4,
						},
						PubKey:            [48]byte{5},
						LastBlockAssigned: 6,
						NextBlockAssigned: 7,
						Score:             8,
						Version:           9,
						Protocol:          10,
						IP:                "test",
						PayeeAddress:      "test2",
					},
				},
			},
		},
	}

	buf := bytes.NewBuffer([]byte{})
	err := state.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var state2 State
	if err := state2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(state2, state); diff != nil {
		t.Fatal(diff)
	}
}