package primitives

// func TestStateDeserializeSerialize(t *testing.T) {
// 	state := State{
// 		UtxoState: UtxoState{
// 			UTXOs: map[chainhash.Hash]Utxo{
// 				chainhash.Hash{
// 					1,
// 				}: {
// 					OutPoint: OutPoint{
// 						chainhash.Hash{
// 							1,
// 						},
// 						2,
// 					},
// 					PrevInputsPubKeys: [][48]byte{
// 						{
// 							3,
// 						},
// 					},
// 					Owner:  "test",
// 					Amount: 4,
// 				},
// 			},
// 		},
// 		GovernanceState: GovernanceState{
// 			Proposals: map[chainhash.Hash]GovernanceProposal{
// 				chainhash.Hash{
// 					14,
// 				}: {
// 					OutPoint: OutPoint{
// 						TxHash: chainhash.Hash{
// 							1,
// 						},
// 						Index: 2,
// 					},
// 					GovData: GovObject{
// 						GovID: chainhash.Hash{
// 							3,
// 						},
// 						Amount:      4,
// 						Cycles:      5,
// 						PayedCycles: 6,
// 						BurnedUtxo: OutPoint{
// 							TxHash: chainhash.Hash{
// 								7,
// 							},
// 							Index: 8,
// 						},
// 						Name:          "test",
// 						URL:           "test2",
// 						PayoutAddress: "test3",
// 						Votes: map[OutPoint]Vote{
// 							OutPoint{
// 								TxHash: chainhash.Hash{
// 									9,
// 								},
// 								Index: 10,
// 							}: {
// 								GovID: chainhash.Hash{
// 									11,
// 								},
// 								Approval: true,
// 								WorkerID: OutPoint{
// 									TxHash: chainhash.Hash{
// 										12,
// 									},
// 									Index: 13,
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		UserState: UserState{
// 			Users: map[chainhash.Hash]User{
// 				chainhash.Hash{
// 					14,
// 				}: {
// 					OutPoint: OutPoint{
// 						TxHash: chainhash.Hash{
// 							1,
// 						},
// 						Index: 2,
// 					},
// 					PubKey: [48]byte{
// 						3,
// 					},
// 					Name: "4",
// 				},
// 			},
// 		},
// 		ValidatorRegistry: []Worker{
// 			{
// 				OutPoint: OutPoint{
// 					TxHash: chainhash.Hash{
// 						1,
// 					},
// 					Index: 2,
// 				},
// 				Balance: 10,
// 				PubKey: [48]byte{
// 					5,
// 				},
// 				PayeeAddress: "test2",
// 			},
// 		},
// 		RANDAO:                        chainhash.Hash{99},
// 		NextRANDAO:                    chainhash.Hash{100},
// 		NextProposerQueue:             []uint32{1, 2, 3},
// 		LatestValidatorRegistryChange: 200,
// 		Slot:                          1,
// 		EpochIndex:                    2,
// 		ProposerQueue:                 []uint32{3},
// 		JustifiedEpoch:                5,
// 		JustifiedEpochHash:            chainhash.Hash{2},
// 		PreviousJustifiedEpochHash:    chainhash.Hash{8, 9},

// 		PreviousJustifiedEpoch: 100,
// 		JustificationBitfield:  6,
// 		FinalizedEpoch:         7,
// 		LatestBlockHashes:      []chainhash.Hash{{8}},
// 		CurrentEpochVotes: []AcceptedVoteInfo{
// 			{
// 				Data: VoteData{
// 					Slot:      3,
// 					FromEpoch: 1,
// 					FromHash:  [32]byte{14},
// 					ToEpoch:   13,
// 					ToHash:    [32]byte{12},
// 				},
// 				ParticipationBitfield: []uint8{1},
// 				Proposer:              5,
// 				InclusionDelay:        3,
// 			},
// 		},
// 		PreviousEpochVotes: []AcceptedVoteInfo{
// 			{
// 				Data: VoteData{
// 					Slot:      3,
// 					FromEpoch: 1,
// 					FromHash:  [32]byte{14},
// 					ToEpoch:   13,
// 					ToHash:    [32]byte{12},
// 				},
// 				ParticipationBitfield: []uint8{1},
// 				Proposer:              5,
// 				InclusionDelay:        3,
// 			},
// 		},
// 	}

// 	buf := bytes.NewBuffer([]byte{})
// 	err := state.Serialize(buf)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	var state2 State
// 	if err := state2.Deserialize(buf); err != nil {
// 		t.Fatal(err)
// 	}

// 	if diff := deep.Equal(state2, state); diff != nil {
// 		t.Fatal(diff)
// 	}
// }
