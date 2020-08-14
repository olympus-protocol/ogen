package mempool_test

import (
	"context"
	"github.com/golang/mock/gomock"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/peers"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

var pool mempool.MockVoteMempool
var mockNet mocknet.Mocknet

func TestNewVoteMempool(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	log := logger.NewMockLogger(ctrl)
	ch := chain.NewMockBlockchain(ctrl)
	host := peers.NewMockHostNode(ctrl)
	host.EXPECT().Topic("votes").Return()
	manager := actionmanager.NewMockLastActionManager(ctrl)

	_, err := mempool.NewVoteMempool(ctx, log, &testdata.IntTestParams, ch, host, manager)
	assert.NoError(t, err)
}
