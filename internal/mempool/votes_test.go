package mempool_test

import (
	"context"
	"github.com/golang/mock/gomock"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/peers"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

var ctx = context.Background()

var mockNet = mocknet.New(ctx)

var pool mempool.VoteMempool

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
	err := pool.AddValidate(&primitives.MultiValidatorVote{}, s)
	assert.NoError(t, err)
}

func TestVoteMempool_Add(t *testing.T) {

}
