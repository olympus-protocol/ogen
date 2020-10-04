package proposer_test

import (
	"context"
	"encoding/hex"
	"github.com/golang/mock/gomock"
	fuzz "github.com/google/gofuzz"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/proposer"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var ctx = context.Background()

var mockNet = mocknet.New(ctx)

var validators []*primitives.Validator
var validatorKeys []*bls.SecretKey

var genesisHash chainhash.Hash

var stateParams initialization.InitializationParameters

func init() {
	f := fuzz.New().NilChance(0)
	f.Fuzz(&genesisHash)
	priv := testdata.PremineAddr

	addrByte, _ := priv.PublicKey().Hash()
	addr := testdata.PremineAddr.PublicKey().ToAccount()

	for i := 0; i < 100; i++ {
		key := bls.RandKey()
		validatorKeys = append(validatorKeys, bls.RandKey())
		val := &primitives.Validator{
			Balance:          100 * 1e8,
			PayeeAddress:     addrByte,
			Status:           primitives.StatusActive,
			FirstActiveEpoch: 0,
			LastActiveEpoch:  0,
		}
		copy(val.PubKey[:], key.PublicKey().Marshal())
		validators = append(validators, val)

	}

	stateParams.GenesisTime = time.Unix(time.Now().Unix(), 0)
	stateParams.InitialValidators = []initialization.ValidatorInitialization{}
	// Convert the validators to initialization params.
	for _, vk := range validatorKeys {
		val := initialization.ValidatorInitialization{
			PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
			PayeeAddress: addr,
		}
		stateParams.InitialValidators = append(stateParams.InitialValidators, val)
	}
	stateParams.PremineAddress = addr
}

func TestProposerWithEmptyKeys(t *testing.T) {
	ctrl := gomock.NewController(t)

	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	genblock := primitives.GetGenesisBlock()

	s := state.NewMockState(ctrl)
	s.EXPECT().GetValidatorRegistry().AnyTimes().Return(validators)
	s.EXPECT().GetVoteCommittee(gomock.Any()).AnyTimes()
	s.EXPECT().GetJustifiedEpoch().Return(uint64(0)).AnyTimes()
	s.EXPECT().GetProposerQueue().Return([]uint64{0, 1, 2, 3, 4, 5}).AnyTimes()
	s.EXPECT().GetJustifiedEpochHash().Return(genblock.Hash()).AnyTimes()
	s.EXPECT().GetRecentBlockHash(gomock.Any()).Return(genblock.Hash()).AnyTimes()

	brow := &chainindex.BlockRow{
		StateRoot: chainhash.Hash{},
		Height:    0,
		Slot:      0,
		Hash:      genblock.Hash(),
		Parent:    nil,
	}

	ch := chain.NewChain(brow)

	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().Add(gomock.Any()).AnyTimes()
	stateService.EXPECT().TipState().AnyTimes().Return(s)
	stateService.EXPECT().TipStateAtSlot(gomock.Any()).Return(s, nil).AnyTimes()
	stateService.EXPECT().Tip().AnyTimes().Return(brow)
	stateService.EXPECT().Chain().Return(ch).AnyTimes()

	bch := chain.NewMockBlockchain(ctrl)
	bch.EXPECT().Notify(gomock.Any())
	bch.EXPECT().State().AnyTimes().Return(stateService)
	bch.EXPECT().GenesisTime().Return(time.Unix(int64(genblock.Header.Timestamp), 0)).AnyTimes()
	bch.EXPECT().Unnotify(gomock.Any())

	host := hostnode.NewMockHostNode(ctrl)
	host.EXPECT().GetHost().Return(h)
	host.EXPECT().Syncing().Return(false).AnyTimes()

	voteMem := mempool.NewMockVoteMempool(ctrl)

	coinsMem := mempool.NewMockCoinsMempool(ctrl)

	actionsMem := mempool.NewMockActionMempool(ctrl)

	am := actionmanager.NewMockLastActionManager(ctrl)
	am.EXPECT().GetNonce().AnyTimes()

	ks := keystore.NewMockKeystore(ctrl)
	ks.EXPECT().OpenKeystore().Times(1)
	ks.EXPECT().GetValidatorKey(gomock.Any()).AnyTimes()

	prop, err := proposer.NewProposer(bch, host, voteMem, coinsMem, actionsMem, am, ks)
	assert.NoError(t, err)
	assert.NotNil(t, prop)

	err = prop.Start()
	assert.NoError(t, err)

	time.Sleep(time.Second * 2)
	prop.Stop()
}

func TestProposerWithKeys(t *testing.T) {
	ctrl := gomock.NewController(t)

	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	genblock := primitives.GetGenesisBlock()

	s := state.NewMockState(ctrl)
	s.EXPECT().GetValidatorRegistry().AnyTimes().Return(validators)
	s.EXPECT().GetVoteCommittee(gomock.Any()).AnyTimes()
	s.EXPECT().GetJustifiedEpoch().Return(uint64(0)).AnyTimes()
	s.EXPECT().GetProposerQueue().Return([]uint64{0, 1, 2, 3, 4, 5}).AnyTimes()
	s.EXPECT().GetJustifiedEpochHash().Return(genblock.Hash()).AnyTimes()
	s.EXPECT().GetRecentBlockHash(gomock.Any()).Return(genblock.Hash()).AnyTimes()

	brow := &chainindex.BlockRow{
		StateRoot: chainhash.Hash{},
		Height:    0,
		Slot:      0,
		Hash:      genblock.Hash(),
		Parent:    nil,
	}

	ch := chain.NewChain(brow)

	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().Add(gomock.Any()).AnyTimes()
	stateService.EXPECT().TipState().AnyTimes().Return(s)
	stateService.EXPECT().TipStateAtSlot(gomock.Any()).Return(s, nil).AnyTimes()
	stateService.EXPECT().Tip().AnyTimes().Return(brow)
	stateService.EXPECT().Chain().Return(ch).AnyTimes()

	bch := chain.NewMockBlockchain(ctrl)
	bch.EXPECT().Notify(gomock.Any())
	bch.EXPECT().State().AnyTimes().Return(stateService)
	bch.EXPECT().GenesisTime().Return(time.Unix(int64(genblock.Header.Timestamp), 0)).AnyTimes()
	bch.EXPECT().Unnotify(gomock.Any())

	host := hostnode.NewMockHostNode(ctrl)
	host.EXPECT().GetHost().Return(h)
	host.EXPECT().Syncing().Return(false).AnyTimes()

	voteMem := mempool.NewMockVoteMempool(ctrl)

	coinsMem := mempool.NewMockCoinsMempool(ctrl)

	actionsMem := mempool.NewMockActionMempool(ctrl)

	am := actionmanager.NewMockLastActionManager(ctrl)
	am.EXPECT().GetNonce().AnyTimes()

	ks := keystore.NewMockKeystore(ctrl)
	ks.EXPECT().OpenKeystore().Times(1)
	for i, v := range validators {
		ks.EXPECT().GetValidatorKey(v.PubKey).Return(validatorKeys[i], true)
	}
	ks.EXPECT().GetValidatorKey(gomock.Any()).AnyTimes()

	prop, err := proposer.NewProposer(bch, host, voteMem, coinsMem, actionsMem, am, ks)
	assert.NoError(t, err)
	assert.NotNil(t, prop)

	err = prop.Start()
	assert.NoError(t, err)
}
