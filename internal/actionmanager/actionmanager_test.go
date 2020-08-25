package actionmanager_test

import (
	"context"
	"github.com/golang/mock/gomock"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

// ctx is the global context used for the entire test
var ctx = context.Background()

// mockNet is a mock network used for PubSubs from libp2p
var mockNet = mocknet.New(ctx)

// genesisHash is just a random hash to set as genesis hash.
var genesisHash chainhash.Hash

// params are the params used on the test
var param = &testdata.TestParams

func TestLastActionManager_New(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Info(gomock.Any()).AnyTimes()

	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	g, err := pubsub.NewGossipSub(ctx, h)
	assert.NoError(t, err)

	bc := chain.NewMockBlockchain(ctrl)
	bc.EXPECT().Notify(gomock.Any())

	host := hostnode.NewMockHostNode(ctrl)
	host.EXPECT().Topic(p2p.MsgValidatorStartCmd).Return(g.Join(p2p.MsgVoteCmd))
	host.EXPECT().GetHost().Return(h)

	am, err := actionmanager.NewLastActionManager(ctx, host, log, bc, param)
	assert.NoError(t, err)
	assert.NotNil(t, am)
}
