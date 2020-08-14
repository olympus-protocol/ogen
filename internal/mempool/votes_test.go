package mempool_test

import (
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
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

type votesTest struct {
	Vote     *primitives.MultiValidatorVote
	Expected error
	Case     string
}

// ctx is the global context used for the entire test
var ctx = context.Background()

// mockNet is a mock network used for PubSubs from libp2p
var mockNet = mocknet.New(ctx)

// pool is a real pool create to test production scenarios
var pool mempool.VoteMempool

// validatorKeys is a slice of signatures that match the validators index
var validatorKeys []*bls.SecretKey

// validators are the initial validators on the realState
var validators []*primitives.Validator

// genesisHash is just a random hash to set as genesis hash.
var genesisHash chainhash.Hash

// realState is a real state to compare production validations.
var realState state.State

// params are the params used on the test
var params = &testdata.IntTestParams

// voteCommitment this mocks the state expected votes by validator index.
// For testing, we just use an slice of 50 validators, vote commitment should be 20% of the total mock registry.
var voteCommitment = []uint64{1, 3, 5, 10, 11, 15, 30, 32, 45, 50}

// tests is an slice of tests the pool should pass.
// Since involves signature, aggregation and secret generation we can't pre-fill them.
var tests []votesTest

func init() {
	f := fuzz.New().NilChance(0)
	f.Fuzz(&genesisHash)

	for i := 0; i < 50; i++ {
		key := bls.RandKey()
		validatorKeys = append(validatorKeys, bls.RandKey())
		val := &primitives.Validator{
			Balance:          100 * 1e8,
			PayeeAddress:     [20]byte{},
			Status:           primitives.StatusActive,
			FirstActiveEpoch: 0,
			LastActiveEpoch:  0,
		}
		copy(val.PubKey[:], key.PublicKey().Marshal())
		validators = append(validators, val)
	}

	realState = state.NewState(primitives.CoinsState{}, validators, genesisHash, params)

	tests = append(tests, votesTest{
		Vote: &primitives.MultiValidatorVote{
			Data:                  nil,
			Sig:                   [96]byte{1, 2, 3},
			ParticipationBitfield: bitfield.NewBitlist(1),
		},
		Expected: state.ErrorVoteEmpty,
		Case:     "Try to validate a vote with null vote data",
	})

	tests = append(tests, votesTest{
		Vote: &primitives.MultiValidatorVote{
			Data: &primitives.VoteData{
				Slot:            1,
				FromEpoch:       0,
				FromHash:        [32]byte{},
				ToEpoch:         0,
				ToHash:          [32]byte{},
				BeaconBlockHash: [32]byte{},
				Nonce:           0,
			},
			Sig:                   [96]byte{1, 2, 3},
			ParticipationBitfield: nil,
		},
		Expected: state.ErrorVoteEmpty,
		Case:     "Try to validate a vote with null bitfield information",
	})

	tests = append(tests, votesTest{
		Vote: &primitives.MultiValidatorVote{
			Data: &primitives.VoteData{
				Slot:            1,
				FromEpoch:       0,
				FromHash:        [32]byte{},
				ToEpoch:         0,
				ToHash:          [32]byte{},
				BeaconBlockHash: [32]byte{},
				Nonce:           0,
			},
			Sig:                   [96]byte{},
			ParticipationBitfield: bitfield.NewBitlist(1),
		},
		Expected: state.ErrorVoteEmpty,
		Case:     "Try to validate a with null signature",
	})

	tests = append(tests, votesTest{
		Vote: &primitives.MultiValidatorVote{
			Data: &primitives.VoteData{
				Slot:            0,
				FromEpoch:       0,
				FromHash:        [32]byte{},
				ToEpoch:         0,
				ToHash:          [32]byte{},
				BeaconBlockHash: [32]byte{},
				Nonce:           0,
			},
			Sig:                   [96]byte{1, 2, 3},
			ParticipationBitfield: bitfield.NewBitlist(1),
		},
		Expected: state.ErrorVoteSlot,
		Case:     "Try to validate a with a wrong slot",
	})

	tests = append(tests, votesTest{
		Vote: &primitives.MultiValidatorVote{
			Data: &primitives.VoteData{
				Slot:            1,
				FromEpoch:       1,
				FromHash:        [32]byte{},
				ToEpoch:         0,
				ToHash:          [32]byte{},
				BeaconBlockHash: [32]byte{},
				Nonce:           0,
			},
			Sig:                   [96]byte{1, 2, 3},
			ParticipationBitfield: bitfield.NewBitlist(1),
		},
		Expected: state.ErrorFromEpoch,
		Case:     "Try to validate a vote that uses a wrong from epoch slot",
	})

	tests = append(tests, votesTest{
		Vote: &primitives.MultiValidatorVote{
			Data: &primitives.VoteData{
				Slot:            1,
				FromEpoch:       0,
				FromHash:        [32]byte{},
				ToEpoch:         0,
				ToHash:          [32]byte{},
				BeaconBlockHash: [32]byte{},
				Nonce:           0,
			},
			Sig:                   [96]byte{1, 2, 3},
			ParticipationBitfield: bitfield.NewBitlist(1),
		},
		Expected: state.ErrorJustifiedHash,
		Case:     "Try to validate a vote uses a wrong from hash",
	})

}

func TestNewVoteMempool(t *testing.T) {

	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	g, err := pubsub.NewGossipSub(ctx, h)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)

	host := peers.NewMockHostNode(ctrl)
	host.EXPECT().Topic("votes").Return(g.Join("votes"))
	host.EXPECT().GetHost().Return(h)

	log := logger.NewMockLogger(ctrl)
	ch := chain.NewMockBlockchain(ctrl)
	manager := actionmanager.NewMockLastActionManager(ctrl)

	pool, err = mempool.NewVoteMempool(ctx, log, &testdata.IntTestParams, ch, host, manager)
	assert.NoError(t, err)
}

func TestVoteMempool_AddValidate(t *testing.T) {
	s := state.NewMockState(gomock.NewController(t))
	s.EXPECT().GetVoteCommittee(10, params).Return(voteCommitment, nil)
	for _, test := range tests {
		s.EXPECT().IsVoteValid(test.Vote, params).Return(realState.IsVoteValid(test.Vote, params))
		err := pool.AddValidate(test.Vote, s)
		assert.Equal(t, test.Expected, err, test.Case)
	}

}

func TestVoteMempool_Add(t *testing.T) {

}
