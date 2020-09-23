package hostnode_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
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

// ctx is the global context used for the entire test
var ctx = context.Background()

// mockNet is a mock network used for PubSubs from libp2p
var mockNet = mocknet.New(ctx)

var mockPeersNum = 25
var mockPeers = make([]host.Host, mockPeersNum)

func init() {
	for i := 0; i < mockPeersNum; i++ {
		p, _ := mockNet.GenPeer()
		mockPeers[i] = p
	}
}

func TestDiscoveryProtocol_New(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Info(gomock.Any()).AnyTimes()

	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	g, err := pubsub.NewGossipSub(ctx, h)
	assert.NoError(t, err)

	bc := chain.NewMockBlockchain(ctrl)
	bc.EXPECT().Notify(gomock.Any())

	db := hostnode.NewMockDatabase(ctrl)
	db.EXPECT().GetSavedPeers().Return([]peer.AddrInfo{}, nil)

	hn := hostnode.NewMockHostNode(ctrl)
	hn.EXPECT().Topic(p2p.MsgValidatorStartCmd).Return(g.Join(p2p.MsgVoteCmd))
	hn.EXPECT().GetHost().Return(h)
	hn.EXPECT().SetStreamHandler(gomock.Any(), gomock.Any()).AnyTimes()
	hn.EXPECT().Notify(gomock.Any()).Times(4)
	hn.EXPECT().Database().Return(db)
	var c hostnode.Config
	c.Log = log
	c.Path = testdata.Node1Folder
	c.Port = testdata.Conf.Port
	c.InitialNodes = testdata.Conf.InitialNodes

	dp, err := hostnode.NewDiscoveryProtocol(ctx, hn, c)
	assert.NoError(t, err)
	assert.NotNil(t, dp)

	err = dp.Start()
	assert.NoError(t, err)
}
