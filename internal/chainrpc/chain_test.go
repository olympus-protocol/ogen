package chainrpc

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	valInfo = state.ValidatorsInfo{
		Validators:  []*primitives.Validator{},
		Active:      5,
		PendingExit: 6,
		PenaltyExit: 7,
		Exited:      8,
		Starting:    9,
	}
	tip = &chainindex.BlockRow{
		StateRoot: chainhash.Hash{1, 2, 3},
		Height:    200,
		Slot:      200,
		Hash:      chainhash.Hash{1, 2, 3},
		Parent:    nil,
	}
)

func Test_ChainServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	s := state.NewMockState(ctrl)
	s.EXPECT().GetValidators().Return(valInfo)

	mockStateService := chain.NewMockStateService(ctrl)
	mockStateService.EXPECT().TipState().Return(s)
	mockStateService.EXPECT().Tip().Return(tip)

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().Return(mockStateService)
	server := chainServer{
		chain:                    ch,
		UnimplementedChainServer: proto.UnimplementedChainServer{},
	}

	info, err := server.GetChainInfo(ctx, &proto.Empty{})

	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, valInfo.Starting, info.Validators.Starting)
	assert.Equal(t, valInfo.Active, info.Validators.Active)
	assert.Equal(t, valInfo.Exited, info.Validators.Exited)
	assert.Equal(t, valInfo.PenaltyExit, info.Validators.PenaltyExit)
	assert.Equal(t, valInfo.PendingExit, info.Validators.PendingExit)
	assert.Equal(t, tip.Hash.String(), info.BlockHash)
	assert.Equal(t, tip.Height, info.BlockHeight)

	//accInfo, err := server.GetAccountInfo(ctx, &proto.Account{})
	//assert.NoError(t, err)
	//assert.NotNil(t, accInfo)

	//block, err := server.GetBlock(ctx, &proto.Hash{})
	//assert.NoError(t, err)
	//assert.NotNil(t, block)

	//hash, err := server.GetBlockHash(ctx, &proto.Number{})
	//assert.NoError(t, err)
	//assert.NotNil(t, hash)

	//rawBlock, err := server.GetRawBlock(ctx, &proto.Hash{})
	//assert.NoError(t, err)
	//assert.NotNil(t, rawBlock)

	//tx, err := server.GetTransaction(ctx, &proto.Hash{})
	//assert.NoError(t, err)
	//assert.NotNil(t, tx)
	//err = server.SubscribeBlocks()
	//assert.NoError(t, err)

	//err = server.SubscribeTransactions()
	//assert.NoError(t, err)

	//err = server.SubscribeValidatorTransaction()
	//assert.NoError(t, err)

	//err = server.Sync()
	//assert.NoError(t, err)

}
