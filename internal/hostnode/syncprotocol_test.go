package hostnode_test

import (
	"github.com/golang/mock/gomock"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSyncProtocol_New(t *testing.T) {
	//f := fuzz.New().NilChance(0)
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Info(gomock.Any()).AnyTimes()

	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	g, err := pubsub.NewGossipSub(ctx, h)
	assert.NoError(t, err)

	host := hostnode.NewMockHostNode(ctrl)
	host.EXPECT().Topic(p2p.MsgBlockCmd).Return(g.Join(p2p.MsgBlocksCmd))
	host.EXPECT().GetHost().Return(h)
	host.EXPECT().SetStreamHandler(gomock.Any(), gomock.Any()).AnyTimes()
	host.EXPECT().Notify(gomock.Any()).Times(4)

	bc := chain.NewMockBlockchain(ctrl)
	bc.EXPECT().Notify(gomock.Any())

	var c hostnode.Config
	c.Log = log
	c.Path = testdata.Node1Folder
	c.Port = testdata.Conf.Port
	c.InitialNodes = testdata.Conf.InitialNodes

	sp, err := hostnode.NewSyncProtocol(ctx, host, c, bc)
	assert.NoError(t, err)
	assert.NotNil(t, sp)

	// func NewDiscoveryProtocol(ctx context.Context, host HostNode, config Config) (DiscoveryProtocol, error) {

}
