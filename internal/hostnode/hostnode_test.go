package hostnode_test

import (
<<<<<<< HEAD
	"context"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
=======
	"github.com/golang/mock/gomock"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
>>>>>>> unit-testing
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

<<<<<<< HEAD
func init() {
	_ = os.Remove("./test")
	_ = os.MkdirAll("./test/hn1/", 0777)
	_ = os.MkdirAll("./test/hn2/", 0777)
}

func TestHostNode(t *testing.T) {

	ctx := context.Background()

	brow := &chainindex.BlockRow{
		StateRoot: chainhash.Hash{},
		Height:    0,
		Slot:      0,
		Hash:      chainhash.Hash{},
		Parent:    nil,
	}

	ctrl := gomock.NewController(t)

	s := state.NewMockState(ctrl)
	s.EXPECT().GetJustifiedEpochHash().Return(chainhash.Hash{}).Times(4)
	s.EXPECT().GetJustifiedEpoch().Return(uint64(1)).Times(4)

	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().Tip().Return(brow).Times(4)
	stateService.EXPECT().TipState().Return(s).Times(4)

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().Return(stateService).Times(6)

	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Info(gomock.Any()).AnyTimes()
	log.EXPECT().Tracef(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Trace(gomock.Any()).AnyTimes()

	cfg := hostnode.Config{
		Log:          log,
		Port:         "50000",
		InitialNodes: nil,
		Path:         "./test/hn1",
		PrivateKey:   nil,
	}

	hn, err := hostnode.NewHostNode(ctx, cfg, ch, testdata.TestParams.NetMagic)
	assert.NoError(t, err)

	cfg.Path = "./test/hn2"
	cfg.Port = "55554"
	hn2, err := hostnode.NewHostNode(ctx, cfg, ch, testdata.TestParams.NetMagic)
	assert.NoError(t, err)

	assert.True(t, hn.Syncing())

	assert.Equal(t, ctx, hn.GetContext())

	assert.Equal(t, testdata.TestParams.NetMagic, hn.GetNetMagic())

	plist := hn.GetPeerList()
	assert.Equal(t, []peer.ID{}, plist)

	pinfo := hn.GetPeerInfos()
	assert.Equal(t, []peer.AddrInfo{}, pinfo)

	npinfo := peer.AddrInfo{
		ID:    hn2.GetHost().ID(),
		Addrs: hn2.GetHost().Addrs(),
	}

	err = hn.GetHost().Connect(ctx, npinfo)
	assert.NoError(t, err)

	assert.True(t, hn.ConnectedToPeer(hn2.GetHost().ID()))

	peers := hn.PeersConnected()
	assert.Equal(t, 1, peers)

	pinfo = hn.GetPeerInfos()
	assert.Equal(t, []peer.AddrInfo{npinfo}, pinfo)

	err = hn.Database().SavePeer(&npinfo)
	assert.NoError(t, err)
=======
func TestHostNode_New(t *testing.T) {

	err := os.Mkdir(testdata.Node1Folder, 0777)
	assert.Nil(t, err)

	//f := fuzz.New().NilChance(0)
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Info(gomock.Any()).AnyTimes()
	log.EXPECT().Infof("binding to address: %s", gomock.Any())

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
	cleanFolder1()

>>>>>>> unit-testing
}
