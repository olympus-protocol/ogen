package primitives

// import (
// 	"bytes"
// 	"testing"
// 	"time"

// 	"github.com/go-test/deep"
// 	"github.com/olympus-protocol/ogen/bls"
// )

// func TestBlockSerializeDeserialize(t *testing.T) {
// 	sig := bls.NewAggregateSignature()
// 	b1 := &Block{
// 		Header: BlockHeader{
// 			Version:       1,
// 			Nonce:         2,
// 			MerkleRoot:    [32]byte{3},
// 			PrevBlockHash: [32]byte{4},
// 			Timestamp:     time.Unix(100, 0),
// 			Slot:          6,
// 		},
// 		Votes: []MultiValidatorVote{
// 			{
// 				Data: VoteData{
// 					Slot:      99,
// 					FromEpoch: 2,
// 					FromHash:  [32]byte{3},
// 					ToEpoch:   4,
// 					ToHash:    [32]byte{5},
// 				},
// 				Signature:             *sig,
// 				ParticipationBitfield: []byte{1, 2, 3, 4},
// 			},
// 		},
// 		StateRoot:       [32]byte{7},
// 		Txs:             []Tx{},
// 		Signature:       [96]byte{8, 9},
// 		RandaoSignature: [96]byte{10, 11},
// 	}

// 	b := bytes.NewBuffer([]byte{})

// 	if err := b1.Encode(b); err != nil {
// 		t.Fatal(err)
// 	}

// 	b2 := new(Block)

// 	if err := b2.Decode(b); err != nil {
// 		t.Fatal(err)
// 	}

// 	if diff := deep.Equal(b1, b2); diff != nil {
// 		t.Fatal(diff)
// 	}
// }
