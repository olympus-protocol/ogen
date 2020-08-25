package hostnode_test

import (
	"context"
	"github.com/golang/mock/gomock"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)


// params are the params used on the test
var param = &testdata.TestParams

// ctx is the global context used for the entire test
var ctx = context.Background()

// mockNet is a mock network used for PubSubs from libp2p
var mockNet = mocknet.New(ctx)

func TestDiscoveryProtocol_New(t *testing.T) {
	//f := fuzz.New().NilChance(0)
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
	host.EXPECT().

	var c hostnode.Config
	c.Log = log
	c.Path = testdata.Node1Folder
	c.Port = testdata.Conf.Port
	c.InitialNodes = testdata.Conf.InitialNodes

	//f.Fuzz(&c.PrivateKey)

	dp, err := hostnode.NewDiscoveryProtocol(ctx, host, c)
	assert.NoError(t, err)
	assert.NotNil(t, dp)

	// func NewDiscoveryProtocol(ctx context.Context, host HostNode, config Config) (DiscoveryProtocol, error) {

}

