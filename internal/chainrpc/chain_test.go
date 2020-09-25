package chainrpc

import (
	"context"
	"encoding/hex"
	"github.com/golang/mock/gomock"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
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

var validatorRegistry []*primitives.Validator
var validatorKeys []*bls.SecretKey

var addrBytes, _ = hex.DecodeString("383e078b9c0089b908050eaa8efd7ba64cbdc5f9ca575e49d9f37b85633f64b6")
var addr, _ = bls.SecretKeyFromBytes(addrBytes)
var addrHash, _ = addr.PublicKey().Hash()

func init() {
	for i := 0; i < 100; i++ {
		key := bls.RandKey()
		validatorKeys = append(validatorKeys, bls.RandKey())
		val := &primitives.Validator{
			Balance:          100 * 1e8,
			PayeeAddress:     addrHash,
			Status:           primitives.StatusActive,
			FirstActiveEpoch: 0,
			LastActiveEpoch:  0,
		}

		copy(val.PubKey[:], key.PublicKey().Marshal())
		validatorRegistry = append(validatorRegistry, val)
	}
}

func Test_ChainServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	cs := primitives.CoinsState{
		Balances: make(map[[20]byte]uint64),
		Nonces:   make(map[[20]byte]uint64),
	}

	genesisBlock := primitives.GetGenesisBlock()
	genesisHash := genesisBlock.Hash()
	genesisBytes, err := genesisBlock.Marshal()
	assert.NoError(t, err)

	s := state.NewMockState(ctrl)
	s.EXPECT().GetValidators().Return(valInfo)
	s.EXPECT().GetCoinsState().Return(cs).AnyTimes()
	s.EXPECT().GetValidatorRegistry().Return(validatorRegistry)
	s.EXPECT().GetFinalizedEpoch().Return(uint64(1))
	s.EXPECT().GetJustifiedEpoch().Return(uint64(1))
	s.EXPECT().GetJustifiedEpochHash().Return(chainhash.Hash{})

	mockStateService := chain.NewMockStateService(ctrl)
	mockStateService.EXPECT().TipState().Return(s).AnyTimes()
	mockStateService.EXPECT().Tip().Return(tip)

	ch := chain.NewMockBlockchain(ctrl)
	ch.EXPECT().State().Return(mockStateService).AnyTimes()
	ch.EXPECT().GetRawBlock(genesisHash).Return(genesisBytes, nil)

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
	assert.Equal(t, uint64(1), info.LastFinalizedEpoch)
	assert.Equal(t, uint64(1), info.LastJustifiedEpoch)
	assert.Equal(t, chainhash.Hash{}.String(), info.LastJustifiedHash)

	block, err := server.GetRawBlock(ctx, &proto.Hash{Hash: genesisHash.String()})
	assert.NoError(t, err)
	assert.NotNil(t, block)

	assert.Equal(t, hex.EncodeToString(genesisBytes), block.RawBlock)

	//accInfo, err := server.GetAccountInfo(ctx, &proto.Account{Account: addr.PublicKey().ToAccount()})
	//assert.NoError(t, err)
	//assert.NotNil(t, accInfo)

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
