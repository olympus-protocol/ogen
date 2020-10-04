package mempool_test

import (
	"github.com/golang/mock/gomock"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCoinsMempool_New(t *testing.T) {
	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)

	host := hostnode.NewMockHostNode(ctrl)
	host.EXPECT().GetHost().Return(h)

	s := state.NewMockState(ctrl)
	s.EXPECT().GetValidatorRegistry().AnyTimes().Return(validatorsGlobal)

	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().TipStateAtSlot(uint64(2)).Times(2).Return(s, nil)
	stateService.EXPECT().TipStateAtSlot(uint64(3)).Times(2).Return(s, nil)

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().AnyTimes().Return(stateService)
	ch.EXPECT().Notify(gomock.Any()).AnyTimes()

	cm, err := mempool.NewCoinsMempool(ch, host)
	assert.NoError(t, err)
	assert.NotNil(t, cm)
}
