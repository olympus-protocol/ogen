package hostnode_test

import (
	"github.com/golang/mock/gomock"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestHostNode_New(t *testing.T) {
	// Create datafolder
	_ = os.Mkdir(testdata.Node1Folder, 0777)

	//f := fuzz.New().NilChance(0)
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Info(gomock.Any()).AnyTimes()

	/*h, err := mockNet.GenPeer()
	assert.NoError(t, err)*/

	/*g, err := pubsub.NewGossipSub(ctx, h)
	assert.NoError(t, err)*/

	bc := chain.NewMockBlockchain(ctrl)
	bc.EXPECT().Notify(gomock.Any())

	var c hostnode.Config
	c.Log = log
	c.Path = testdata.Node1Folder
	c.Port = testdata.Conf.Port
	c.InitialNodes = testdata.Conf.InitialNodes

	host, err := hostnode.NewHostNode(ctx, c, bc)
	assert.NoError(t, err)
	assert.NotNil(t, host)

}
