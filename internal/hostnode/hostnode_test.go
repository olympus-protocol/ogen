package hostnode_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
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

	//mockNet := mocknet.New(ctx)

	brow := &chainindex.BlockRow{
		StateRoot: chainhash.Hash{},
		Height:    0,
		Slot:      0,
		Hash:      chainhash.Hash{},
		Parent:    nil,
	}

	ctrl := gomock.NewController(t)
	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().Tip().Return(brow).Times(2)

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().Return(stateService).Times(2)

	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Info(gomock.Any()).AnyTimes()
	log.EXPECT().Tracef(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Trace(gomock.Any()).AnyTimes()

	cfg := hostnode.Config{
		Log:          log,
		Port:         "55555",
		InitialNodes: nil,
		Path:         "./test/hn1",
		PrivateKey:   nil,
	}

	hn, err := hostnode.NewHostNode(ctx, cfg, ch, testdata.TestParams.NetMagic)
	assert.NoError(t, err)

	cfg.Path = "./test/hn2"
	cfg.Port = "55555"
	hn2, err := hostnode.NewHostNode(ctx, cfg, ch, testdata.TestParams.NetMagic)
	assert.NoError(t, err)

	assert.False(t, hn.Syncing())

	assert.Equal(t, ctx, hn.GetContext())

	assert.Equal(t, testdata.TestParams.NetMagic, hn.GetNetMagic())

	plist := hn.GetPeerList()
	assert.Equal(t, []peer.ID{}, plist)

	pinfo := hn.GetPeerInfos()
	assert.Equal(t, []peer.AddrInfo{}, pinfo)

	_ = peer.AddrInfo{
		ID:    hn2.GetHost().ID(),
		Addrs: hn2.GetHost().Addrs(),
	}

	//err = hn.GetHost().Connect(ctx, npinfo)
	assert.NoError(t, err)
}
