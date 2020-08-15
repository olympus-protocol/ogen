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
	"time"
)

type votesTest struct {
	Vote *primitives.MultiValidatorVote
}

// ctx is the global context used for the entire test
var ctx = context.Background()

// mockNet is a mock network used for PubSubs from libp2p
var mockNet = mocknet.New(ctx)

// validatorKeys is a slice of signatures that match the validators index
var validatorKeys []*bls.SecretKey

// validators are the initial validators on the realState
var validators []*primitives.Validator

// genesisHash is just a random hash to set as genesis hash.
var genesisHash chainhash.Hash

// params are the params used on the test
var params = &testdata.IntTestParams

// voteCommitment this mocks the state expected votes by validator index.
// For testing, we just use an slice of 50 validators, vote commitment should be 20% of the total mock registry.
var voteCommitment = []uint64{1, 3, 5, 10, 11, 15, 30, 32, 45, 49}

var goodVotes []votesTest

var slashedVotes []votesTest

var slot uint64 = 10
var dataNonce uint64 = 10
var aggVote *primitives.MultiValidatorVote

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

	voteData := &primitives.VoteData{
		Slot:            slot,
		FromEpoch:       0,
		FromHash:        genesisHash,
		ToEpoch:         11,
		ToHash:          genesisHash,
		BeaconBlockHash: [32]byte{},
		Nonce:           dataNonce,
	}

	voteDataSlash := &primitives.VoteData{
		Slot:            slot,
		FromEpoch:       0,
		FromHash:        genesisHash,
		ToEpoch:         11,
		ToHash:          genesisHash,
		BeaconBlockHash: [32]byte{1,2,3},
		Nonce:           dataNonce,
	}

	voteDataHash := voteData.Hash()
	voteDataSlashHash := voteDataSlash.Hash()

	var sigs1 [][96]byte
	var sigs2 [][96]byte
	var sigsSlash1 [][96]byte
	var sigsSlash2 [][96]byte

	bf1 := bitfield.NewBitlist(uint64(len(voteCommitment)))
	bf2 := bitfield.NewBitlist(uint64(len(voteCommitment)))
	bfSlash := bitfield.NewBitlist(uint64(len(voteCommitment)))

	for i, idx := range voteCommitment {

		if i == 0 || i == 2 || i == 4 || i == 7 || i == 9 {

			key := validatorKeys[idx]
			bfSlash.Set(uint(i))

			var sigB1 [96]byte
			sig1 := key.Sign(voteDataHash[:]).Marshal()
			copy(sigB1[:], sig1)
			sigsSlash1 = append(sigsSlash1, sigB1)

			var sigB2 [96]byte
			sig2 := key.Sign(voteDataSlashHash[:]).Marshal()
			copy(sigB2[:], sig2)
			sigsSlash2 = append(sigsSlash2, sigB2)
		}

		if i < 5 {
			key := validatorKeys[idx]
			bf1.Set(uint(i))
			var sigB [96]byte
			sig := key.Sign(voteDataHash[:]).Marshal()
			copy(sigB[:], sig)
			sigs1 = append(sigs1, sigB)
		} else {
			key := validatorKeys[idx]
			bf2.Set(uint(i))
			var sigB [96]byte
			sig := key.Sign(voteDataHash[:]).Marshal()
			copy(sigB[:], sig)
			sigs2 = append(sigs2, sigB)
		}

	}

	var sigB1 [96]byte
	var sigB2 [96]byte
	var sigSlashB1 [96]byte
	var sigSlashB2 [96]byte

	sig1, err := bls.AggregateSignaturesBytes(sigs1)
	if err != nil {
		panic(err)
	}
	sig2, err := bls.AggregateSignaturesBytes(sigs2)
	if err != nil {
		panic(err)
	}
	sigSlash1, err := bls.AggregateSignaturesBytes(sigsSlash1)
	if err != nil {
		panic(err)
	}
	sigSlash2, err := bls.AggregateSignaturesBytes(sigsSlash2)
	if err != nil {
		panic(err)
	}

	copy(sigB1[:], sig1.Marshal())
	copy(sigB2[:], sig2.Marshal())
	copy(sigSlashB1[:], sigSlash1.Marshal())
	copy(sigSlashB2[:], sigSlash2.Marshal())

	goodVotes = append(goodVotes, votesTest{
		Vote: &primitives.MultiValidatorVote{
			Data:                  voteData,
			Sig:                   sigB1,
			ParticipationBitfield: bf1,
		},
	})

	goodVotes = append(goodVotes, votesTest{
		Vote: &primitives.MultiValidatorVote{
			Data:                  voteData,
			Sig:                   sigB2,
			ParticipationBitfield: bf2,
		},
	})

	slashedVotes = append(slashedVotes, votesTest{
		Vote: &primitives.MultiValidatorVote{
			Data:                  voteData,
			Sig:                   sigSlashB1,
			ParticipationBitfield: bfSlash,
		},
	})

	slashedVotes = append(slashedVotes, votesTest{
		Vote: &primitives.MultiValidatorVote{
			Data:                  voteDataSlash,
			Sig:                   sigSlashB2,
			ParticipationBitfield: bfSlash,
		},
	})

	aggSig, err := bls.AggregateSignaturesBytes([][96]byte{sigB1, sigB2})
	if err != nil {
		panic(err)
	}

	var aggSigBytes [96]byte
	copy(aggSigBytes[:], aggSig.Marshal())

	mbf, err := bf1.Merge(bf2)
	if err != nil {
		panic(err)
	}

	aggVote = &primitives.MultiValidatorVote{
		Data:                  voteData,
		Sig:                   aggSigBytes,
		ParticipationBitfield: mbf,
	}
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
	s.EXPECT().GetVoteCommittee(slot, params).AnyTimes().Return(voteCommitment, nil)
	s.EXPECT().GetValidatorRegistry().AnyTimes().Return(validators)
	s.EXPECT().ProcessVote(aggVote, params, uint64(1)).Return(nil)

	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().TipStateAtSlot(slot+params.MinAttestationInclusionDelay).AnyTimes().Return(s, nil)

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().AnyTimes().Return(stateService)

	manager := actionmanager.NewMockLastActionManager(ctrl)
	for _, i := range voteCommitment {
		manager.EXPECT().RegisterAction(validators[i].PubKey, dataNonce).AnyTimes()
	}

	pool, err := mempool.NewVoteMempool(ctx, log, &testdata.IntTestParams, ch, host, manager)
	assert.NoError(t, err)

	for _, test := range goodVotes {
		s.EXPECT().IsVoteValid(test.Vote, params).Return(nil)
		err := pool.AddValidate(test.Vote, s)
		assert.NoError(t, err)
	}

	votes, err := pool.Get(slot+1, s, params, 1)
	assert.NoError(t, err)
	assert.Equal(t, aggVote, votes[0])
	assert.Equal(t, 1, len(votes))
}

func TestVoteMempoolSlashing1(t *testing.T) {

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
	s.EXPECT().GetVoteCommittee(slot, params).AnyTimes().Return(voteCommitment, nil)
	s.EXPECT().GetValidatorRegistry().AnyTimes().Return(validators)
	s.EXPECT().ProcessVote(goodVotes[0].Vote, params, uint64(1)).Return(nil)

	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().TipStateAtSlot(slot+params.MinAttestationInclusionDelay).AnyTimes().Return(s, nil)

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().AnyTimes().Return(stateService)

	manager := actionmanager.NewMockLastActionManager(ctrl)
	for _, i := range voteCommitment {
		manager.EXPECT().RegisterAction(validators[i].PubKey, dataNonce).AnyTimes()
	}

	pool, err := mempool.NewVoteMempool(ctx, log, &testdata.IntTestParams, ch, host, manager)
	assert.NoError(t, err)

	slashNotify := mempool.NewMockVoteSlashingNotifee(ctrl)
	slashNotify.EXPECT().NotifyIllegalVotes(&primitives.VoteSlashing{
		Vote1: slashedVotes[0].Vote,
		Vote2: goodVotes[0].Vote,
	})
	pool.Notify(slashNotify)

	// First we submit the first vote twice to confirm that mempool rejects equal votes
	s.EXPECT().IsVoteValid(goodVotes[0].Vote, params).Return(nil)
	err = pool.AddValidate(goodVotes[0].Vote, s)
	assert.NoError(t, err)

	s.EXPECT().IsVoteValid(slashedVotes[0].Vote, params).Return(nil)
	err = pool.AddValidate(slashedVotes[0].Vote, s)
	assert.NoError(t, err)

	votes, err := pool.Get(slot+1, s, params, 1)
	assert.NoError(t, err)
	assert.Equal(t, goodVotes[0].Vote, votes[0])
	assert.Equal(t, 1, len(votes))
}

func TestVoteMempoolSlashing2(t *testing.T) {

	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	g, err := pubsub.NewGossipSub(ctx, h)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)

	host := peers.NewMockHostNode(ctrl)
	host.EXPECT().Topic("votes").Return(g.Join("votes"))
	host.EXPECT().GetHost().Return(h)

	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Warnf("found surround vote for multivalidator in vote %s ...", slashedVotes[1].Vote.Data.String())

	s := state.NewMockState(ctrl)
	s.EXPECT().GetVoteCommittee(slot, params).AnyTimes().Return(voteCommitment, nil)
	s.EXPECT().GetValidatorRegistry().AnyTimes().Return(validators)
	s.EXPECT().ProcessVote(goodVotes[0].Vote, params, uint64(1)).Return(nil)

	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().TipStateAtSlot(slot+params.MinAttestationInclusionDelay).AnyTimes().Return(s, nil)

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().AnyTimes().Return(stateService)

	manager := actionmanager.NewMockLastActionManager(ctrl)
	for _, i := range voteCommitment {
		manager.EXPECT().RegisterAction(validators[i].PubKey, dataNonce).AnyTimes()
	}

	pool, err := mempool.NewVoteMempool(ctx, log, &testdata.IntTestParams, ch, host, manager)
	assert.NoError(t, err)

	slashNotify := mempool.NewMockVoteSlashingNotifee(ctrl)
	slashNotify.EXPECT().NotifyIllegalVotes(&primitives.VoteSlashing{
		Vote1: slashedVotes[1].Vote,
		Vote2: goodVotes[0].Vote,
	})
	pool.Notify(slashNotify)

	// First we submit the first vote twice to confirm that mempool rejects equal votes
	s.EXPECT().IsVoteValid(goodVotes[0].Vote, params).Return(nil)
	err = pool.AddValidate(goodVotes[0].Vote, s)
	assert.NoError(t, err)

	s.EXPECT().IsVoteValid(slashedVotes[1].Vote, params).Return(nil)
	err = pool.AddValidate(slashedVotes[1].Vote, s)
	assert.NoError(t, err)

	votes, err := pool.Get(slot+1, s, params, 1)
	assert.NoError(t, err)
	assert.Equal(t, goodVotes[0].Vote, votes[0])
	assert.Equal(t, 1, len(votes))
}

func TestVoteMempoolRelayed(t *testing.T) {

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
	s.EXPECT().GetVoteCommittee(slot, params).AnyTimes().Return(voteCommitment, nil)
	s.EXPECT().GetValidatorRegistry().AnyTimes().Return(validators)
	s.EXPECT().ProcessVote(goodVotes[0].Vote, params, uint64(1)).Return(nil)

	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().TipStateAtSlot(slot+params.MinAttestationInclusionDelay).AnyTimes().Return(s, nil)

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().AnyTimes().Return(stateService)

	manager := actionmanager.NewMockLastActionManager(ctrl)
	for _, i := range voteCommitment {
		manager.EXPECT().RegisterAction(validators[i].PubKey, dataNonce).AnyTimes()
	}

	pool, err := mempool.NewVoteMempool(ctx, log, &testdata.IntTestParams, ch, host, manager)
	assert.NoError(t, err)

	slashNotify := mempool.NewMockVoteSlashingNotifee(ctrl)
	pool.Notify(slashNotify)

	// First we submit the first vote twice to confirm that mempool rejects equal votes
	s.EXPECT().IsVoteValid(goodVotes[0].Vote, params).Return(nil)
	err = pool.AddValidate(goodVotes[0].Vote, s)
	assert.NoError(t, err)

	s.EXPECT().IsVoteValid(goodVotes[0].Vote, params).Return(nil)
	err = pool.AddValidate(goodVotes[0].Vote, s)
	assert.NoError(t, err)
	time.Sleep(time.Second * 1)
	votes, err := pool.Get(slot+1, s, params, 1)
	assert.NoError(t, err)
	assert.Equal(t, goodVotes[0].Vote, votes[0])
	assert.Equal(t, 1, len(votes))
}
