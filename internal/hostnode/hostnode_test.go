package hostnode_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

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

	chview := chain.NewChain(brow)

	s := state.NewMockState(ctrl)
	s.EXPECT().GetJustifiedEpochHash().Return(chainhash.Hash{}).Times(4)
	s.EXPECT().GetJustifiedEpoch().Return(uint64(1)).Times(4)


	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().Tip().Return(brow).Times(4)
	stateService.EXPECT().TipState().Return(s).Times(4)
	stateService.EXPECT().GetFinalizedHead().Return(brow, nil).AnyTimes()
	stateService.EXPECT().GetJustifiedHead().Return(brow, nil).AnyTimes()
	stateService.EXPECT().Chain().Return(chview).AnyTimes()

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().Return(stateService).AnyTimes()

	log := logger.New(os.Stdin)

	cfg := hostnode.Config{
		Log:  log,
		Port: "50000",
		Path: "./test/hn1",
	}

	hn, err := hostnode.NewHostNode(ctx, cfg, ch, &testdata.TestParams)
	assert.NoError(t, err)

	cfg.Path = "./test/hn2"
	cfg.Port = "55554"
	hn2, err := hostnode.NewHostNode(ctx, cfg, ch, &testdata.TestParams)
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

	err = hn.GetHost().Connect(hn.GetContext(), npinfo)
	assert.NoError(t, err)

	assert.True(t, hn.ConnectedToPeer(hn2.GetHost().ID()))

	peers := hn.PeersConnected()
	assert.Equal(t, 1, peers)

	//	pinfo = hn.GetPeerInfos()
	//	assert.Equal(t, pstore_pb.ProtoAddr{Multiaddr: npinfo.Addrs[0]}, pinfo)

	err = hn.Start()
	assert.NoError(t, err)

	err = hn2.Start()
	assert.NoError(t, err)

	_ = os.RemoveAll("./test")
}
