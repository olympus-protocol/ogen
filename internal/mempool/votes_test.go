package mempool_test

import (
	"bytes"
	"context"
	"github.com/golang/mock/gomock"
	fuzz "github.com/google/gofuzz"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/peers"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

// ctx is the global context used for the entire test
var ctx = context.Background()

// mockNet is a mock network used for PubSubs from libp2p
var mockNet = mocknet.New(ctx)

// validatorKeys is a slice of signatures that match the validators index
var validatorKeys1 []bls_interface.SecretKey
var validatorKeys2 []bls_interface.SecretKey

// validators are the initial validators on the realState
var validators1 []*primitives.Validator
var validators2 []*primitives.Validator
var validatorsGlobal []*primitives.Validator

// genesisHash is just a random hash to set as genesis hash.
var genesisHash chainhash.Hash

// params are the params used on the test
var param = &testdata.IntTestParams

var slot1Commiters = []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99}
var slot2Commiters = []uint64{20, 21, 22, 23, 24, 25, 26, 27, 28, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79}
var slot3Commiters = []uint64{40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59}
var slot4Commiters = []uint64{60, 61, 62, 63, 64, 65, 66, 67, 68, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39}
var slot5Commiters = []uint64{80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}

func init() {
	f := fuzz.New().NilChance(0)
	f.Fuzz(&genesisHash)

	for i := 0; i < 100; i++ {
		if i < 50 {
			key := bls.CurrImplementation.RandKey()
			validatorKeys1 = append(validatorKeys1, bls.CurrImplementation.RandKey())
			val := &primitives.Validator{
				Balance:          100 * 1e8,
				PayeeAddress:     [20]byte{},
				Status:           primitives.StatusActive,
				FirstActiveEpoch: 0,
				LastActiveEpoch:  0,
			}
			copy(val.PubKey[:], key.PublicKey().Marshal())
			validators1 = append(validators1, val)
		} else {
			key := bls.CurrImplementation.RandKey()
			validatorKeys2 = append(validatorKeys2, bls.CurrImplementation.RandKey())
			val := &primitives.Validator{
				Balance:          100 * 1e8,
				PayeeAddress:     [20]byte{},
				Status:           primitives.StatusActive,
				FirstActiveEpoch: 0,
				LastActiveEpoch:  0,
			}
			copy(val.PubKey[:], key.PublicKey().Marshal())
			validators2 = append(validators2, val)
		}

	}
	validatorsGlobal = append(validators1, validators2...)
}

func TestVoteMempoolAggregation(t *testing.T) {

	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	g, err := pubsub.NewGossipSub(ctx, h)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)

	host := peers.NewMockHostNode(ctrl)
	host.EXPECT().Topic("votes").Return(g.Join("votes"))
	host.EXPECT().GetHost().Return(h)

	log := logger.NewMockLogger(ctrl)

	s := state.NewMockState(ctrl)
	s.EXPECT().GetValidatorRegistry().AnyTimes().Return(validatorsGlobal)

	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().TipStateAtSlot(uint64(2)).Times(2).Return(s, nil)
	stateService.EXPECT().TipStateAtSlot(uint64(3)).Times(2).Return(s, nil)

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().AnyTimes().Return(stateService)

	manager := actionmanager.NewMockLastActionManager(ctrl)

	pool, err := mempool.NewVoteMempool(ctx, log, &testdata.IntTestParams, ch, host, manager)
	assert.NoError(t, err)

	// This test will try to replicate a chain with 100 validators and 2 proposers moving for 1 epoch.
	slotToVote := uint64(1)

	voteDataSlot1 := &primitives.VoteData{
		Slot:            slotToVote,
		FromEpoch:       0,
		FromHash:        genesisHash,
		ToEpoch:         5,
		ToHash:          [32]byte{},
		BeaconBlockHash: [32]byte{},
		Nonce:           0,
	}

	voteDataSlot1Hash := voteDataSlot1.Hash()

	bfS1att1 := bitfield.NewBitlist(uint64(len(slot1Commiters)))
	bfS1att2 := bitfield.NewBitlist(uint64(len(slot1Commiters)))
	bfS1Aggr := bitfield.NewBitlist(uint64(len(slot1Commiters)))

	var sigsS1Att1 []bls_interface.Signature
	var sigsS1Att2 []bls_interface.Signature
	var sigsS1Aggr []bls_interface.Signature

	for i, val := range slot1Commiters {
		bfS1Aggr.Set(uint(i))

		votingValidator := validatorsGlobal[val]

		manager.EXPECT().RegisterAction(votingValidator.PubKey, uint64(0))

		for j, valAtt1 := range validators1 {
			if bytes.Equal(votingValidator.PubKey[:], valAtt1.PubKey[:]) {
				bfS1att1.Set(uint(i))
				sig := validatorKeys1[j].Sign(voteDataSlot1Hash[:])
				sigsS1Att1 = append(sigsS1Att1, sig)
				sigsS1Aggr = append(sigsS1Aggr, sig)
			}
		}

		for j, valAtt2 := range validators2 {
			if bytes.Equal(votingValidator.PubKey[:], valAtt2.PubKey[:]) {
				bfS1att2.Set(uint(i))
				sig := validatorKeys2[j].Sign(voteDataSlot1Hash[:])
				sigsS1Att2 = append(sigsS1Att2, sig)
				sigsS1Aggr = append(sigsS1Aggr, sig)
			}
		}
	}

	var SigS1Att1 [96]byte
	var SigS1Att2 [96]byte
	var SigS1Aggr [96]byte

	sigS1Att1 := bls.CurrImplementation.AggregateSignatures(sigsS1Att1)
	sigS1Att2 := bls.CurrImplementation.AggregateSignatures(sigsS1Att2)
	sigS1Aggr := bls.CurrImplementation.AggregateSignatures(sigsS1Aggr)

	copy(SigS1Att1[:], sigS1Att1.Marshal())
	copy(SigS1Att2[:], sigS1Att2.Marshal())
	copy(SigS1Aggr[:], sigS1Aggr.Marshal())

	voteSlot1att1 := &primitives.MultiValidatorVote{
		Data:                  voteDataSlot1,
		Sig:                   SigS1Att1,
		ParticipationBitfield: bfS1att1,
	}

	voteSlot1att2 := &primitives.MultiValidatorVote{
		Data:                  voteDataSlot1,
		Sig:                   SigS1Att2,
		ParticipationBitfield: bfS1att2,
	}

	voteSlot1AggVote := &primitives.MultiValidatorVote{
		Data:                  voteDataSlot1,
		Sig:                   SigS1Aggr,
		ParticipationBitfield: bfS1Aggr,
	}

	s.EXPECT().IsVoteValid(voteSlot1att1, param).Return(nil)
	s.EXPECT().IsVoteValid(voteSlot1att2, param).Return(nil)
	s.EXPECT().GetVoteCommittee(voteDataSlot1.Slot, param).AnyTimes().Return(slot1Commiters, nil)
	s.EXPECT().ProcessVote(voteSlot1AggVote, param, uint64(1)).Return(nil)

	err = pool.AddValidate(voteSlot1att1, s)
	assert.NoError(t, err)
	err = pool.AddValidate(voteSlot1att2, s)
	assert.NoError(t, err)

	slotToVote++

	votesSlot1, err := pool.Get(slotToVote, s, param, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(votesSlot1))

	block := &primitives.Block{
		Votes: votesSlot1,
	}

	pool.Remove(block)

	voteDataSlot2 := &primitives.VoteData{
		Slot:            slotToVote,
		FromEpoch:       0,
		FromHash:        genesisHash,
		ToEpoch:         5,
		ToHash:          [32]byte{},
		BeaconBlockHash: [32]byte{},
		Nonce:           0,
	}

	voteDataSlot2Hash := voteDataSlot2.Hash()

	bfs2att1 := bitfield.NewBitlist(uint64(len(slot1Commiters)))
	bfs2att2 := bitfield.NewBitlist(uint64(len(slot1Commiters)))
	bfs2aggr := bitfield.NewBitlist(uint64(len(slot1Commiters)))

	var sigsS2Att1 []bls_interface.Signature
	var sigsS2Att2 []bls_interface.Signature
	var sigsS2Aggr []bls_interface.Signature

	for i, val := range slot2Commiters {
		bfs2aggr.Set(uint(i))

		votingValidator := validatorsGlobal[val]

		manager.EXPECT().RegisterAction(votingValidator.PubKey, uint64(0))

		for j, valAtt1 := range validators1 {
			if bytes.Equal(votingValidator.PubKey[:], valAtt1.PubKey[:]) {
				bfs2att1.Set(uint(i))
				sig := validatorKeys1[j].Sign(voteDataSlot2Hash[:])
				sigsS2Att1 = append(sigsS2Att1, sig)
				sigsS2Aggr = append(sigsS2Aggr, sig)
			}
		}

		for j, valAtt2 := range validators2 {
			if bytes.Equal(votingValidator.PubKey[:], valAtt2.PubKey[:]) {
				bfs2att2.Set(uint(i))
				sig := validatorKeys2[j].Sign(voteDataSlot2Hash[:])
				sigsS2Att2 = append(sigsS2Att2, sig)
				sigsS2Aggr = append(sigsS2Aggr, sig)
			}
		}
	}

	var SigS2Att1 [96]byte
	var SigS2Att2 [96]byte
	var SigS2Aggr [96]byte

	sigS2Att1 := bls.CurrImplementation.AggregateSignatures(sigsS2Att1)
	sigS2Att2 := bls.CurrImplementation.AggregateSignatures(sigsS2Att2)
	sigS2Aggr := bls.CurrImplementation.AggregateSignatures(sigsS2Aggr)

	copy(SigS2Att1[:], sigS2Att1.Marshal())
	copy(SigS2Att2[:], sigS2Att2.Marshal())
	copy(SigS2Aggr[:], sigS2Aggr.Marshal())

	voteSlot2att1 := &primitives.MultiValidatorVote{
		Data:                  voteDataSlot2,
		Sig:                   SigS2Att1,
		ParticipationBitfield: bfs2att1,
	}

	voteSlot2att2 := &primitives.MultiValidatorVote{
		Data:                  voteDataSlot2,
		Sig:                   SigS2Att2,
		ParticipationBitfield: bfs2att2,
	}

	voteSlot2AggVote := &primitives.MultiValidatorVote{
		Data:                  voteDataSlot2,
		Sig:                   SigS2Aggr,
		ParticipationBitfield: bfs2aggr,
	}

	s.EXPECT().IsVoteValid(voteSlot2att1, param).Return(nil)
	s.EXPECT().IsVoteValid(voteSlot2att2, param).Return(nil)
	s.EXPECT().GetVoteCommittee(voteDataSlot2.Slot, param).AnyTimes().Return(slot2Commiters, nil)
	s.EXPECT().ProcessVote(voteSlot2AggVote, param, uint64(1)).Return(nil)

	err = pool.AddValidate(voteSlot2att1, s)
	assert.NoError(t, err)
	err = pool.AddValidate(voteSlot2att2, s)
	assert.NoError(t, err)

	slotToVote++

	votesSlot2, err := pool.Get(slotToVote, s, param, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(votesSlot2))

	blockS2 := &primitives.Block{
		Votes: votesSlot2,
	}

	pool.Remove(blockS2)
}
