package mempool_test

import (
	"github.com/golang/mock/gomock"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCoinsMempool_New(t *testing.T) {
	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	g, err := pubsub.NewGossipSub(ctx, h)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)

	host := hostnode.NewMockHostNode(ctrl)
	host.EXPECT().Topic(p2p.MsgTxCmd).Return(g.Join(p2p.MsgTxCmd))
	host.EXPECT().Topic(p2p.MsgTxMultiCmd).Return(g.Join(p2p.MsgTxMultiCmd))
	host.EXPECT().GetHost().Return(h)

	log := logger.NewMockLogger(ctrl)

	s := state.NewMockState(ctrl)
	s.EXPECT().GetValidatorRegistry().AnyTimes().Return(validatorsGlobal)

	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().TipStateAtSlot(uint64(2)).Times(2).Return(s, nil)
	stateService.EXPECT().TipStateAtSlot(uint64(3)).Times(2).Return(s, nil)

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().AnyTimes().Return(stateService)
	ch.EXPECT().Notify(gomock.Any()).AnyTimes()

	cm, err := mempool.NewCoinsMempool(ctx, log, ch, host, param)
	assert.NoError(t, err)
	assert.NotNil(t, cm)
}
