package proposer_test

import (
	"context"
	"encoding/hex"
	"github.com/golang/mock/gomock"
	fuzz "github.com/google/gofuzz"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/proposer"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// ctx is the global context used for the entire test
var ctx = context.Background()

// mockNet is a mock network used for PubSubs from libp2p
var mockNet = mocknet.New(ctx)

// validatorKeys is a slice of signatures that match the validators index
var validatorKeys1 []bls_interface.SecretKey
var validatorKeys2 []bls_interface.SecretKey

// validators are the initial validators on the realState
var validators1 []*primitives.Validator
var validators2 []*primitives.Validator
var validatorsGlobal []*primitives.Validator

// genesisHash is just a random hash to set as genesis hash.
var genesisHash chainhash.Hash

// params are the params used on the test
var param = &testdata.TestParams

// init params used on the test
var stateParams state.InitializationParameters

func init() {
	f := fuzz.New().NilChance(0)
	f.Fuzz(&genesisHash)
	priv := testdata.PremineAddr

	addrByte, _ := priv.PublicKey().Hash()
	addr := testdata.PremineAddr.PublicKey().ToAccount()

	for i := 0; i < 100; i++ {
		if i < 50 {
			key := bls.CurrImplementation.RandKey()
			validatorKeys1 = append(validatorKeys1, bls.CurrImplementation.RandKey())
			val := &primitives.Validator{
				Balance:          100 * 1e8,
				PayeeAddress:     addrByte,
				Status:           primitives.StatusActive,
				FirstActiveEpoch: 0,
				LastActiveEpoch:  0,
			}
			copy(val.PubKey[:], key.PublicKey().Marshal())
			validators1 = append(validators1, val)
		} else {
			key := bls.CurrImplementation.RandKey()
			validatorKeys2 = append(validatorKeys2, bls.CurrImplementation.RandKey())
			val := &primitives.Validator{
				Balance:          100 * 1e8,
				PayeeAddress:     addrByte,
				Status:           primitives.StatusActive,
				FirstActiveEpoch: 0,
				LastActiveEpoch:  0,
			}
			copy(val.PubKey[:], key.PublicKey().Marshal())
			validators2 = append(validators2, val)
		}

	}
	validatorsGlobal = append(validators1, validators2...)
	stateParams.GenesisTime = time.Unix(time.Now().Unix(), 0)
	stateParams.InitialValidators = []state.ValidatorInitialization{}
	// Convert the validators to initialization params.
	for _, vk := range validatorKeys1 {
		val := state.ValidatorInitialization{
			PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
			PayeeAddress: addr,
		}
		stateParams.InitialValidators = append(stateParams.InitialValidators, val)
	}
	stateParams.PremineAddress = addr
}

// create a blockchain instance and test its methods
func TestProposer_Object(t *testing.T) {
	//f := fuzz.New().NilChance(0)
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Infof("starting proposer with %d/%d active validators", gomock.Any(), gomock.Any())
	log.EXPECT().Infof("sending votes for slot %d", gomock.Any()).AnyTimes()
	log.EXPECT().Debugf("committing for slot %d with %d validators", gomock.Any(), gomock.Any()).AnyTimes()

	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	g, err := pubsub.NewGossipSub(ctx, h)
	assert.NoError(t, err)

	genblock := primitives.GetGenesisBlock()

	s := state.NewMockState(ctrl)
	s.EXPECT().GetValidatorRegistry().AnyTimes().Return(validatorsGlobal)
	s.EXPECT().GetVoteCommittee(gomock.Any(), gomock.Any()).AnyTimes()
	s.EXPECT().GetJustifiedEpoch().Return(uint64(0)).AnyTimes()
	s.EXPECT().GetProposerQueue().Return([]uint64{0, 1, 2, 3, 4, 5}).AnyTimes()
	s.EXPECT().GetJustifiedEpochHash().Return(genblock.Hash()).AnyTimes()
	s.EXPECT().GetRecentBlockHash(gomock.Any(), gomock.Any()).Return(genblock.Hash()).AnyTimes()

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
	stateService.EXPECT().TipState().Times(1).Return(s)
	stateService.EXPECT().TipStateAtSlot(gomock.Any()).Return(s, nil).AnyTimes()
	stateService.EXPECT().Tip().AnyTimes().Return(brow)
	stateService.EXPECT().Chain().Return(ch).AnyTimes()

	bc := chain.NewMockBlockchain(ctrl)
	bc.EXPECT().Notify(gomock.Any())
	bc.EXPECT().State().AnyTimes().Return(stateService)
	bc.EXPECT().GenesisTime().Return(time.Unix(int64(genblock.Header.Timestamp), 0)).AnyTimes()

	host := hostnode.NewMockHostNode(ctrl)
	host.EXPECT().Topic(p2p.MsgBlockCmd).Return(g.Join(p2p.MsgBlockCmd))
	host.EXPECT().Topic(p2p.MsgVoteCmd).Return(g.Join(p2p.MsgVoteCmd))
	host.EXPECT().GetHost().Return(h)
	host.EXPECT().PeersConnected().Return(1).AnyTimes()
	host.EXPECT().Syncing().Return(false).AnyTimes()

	voteMem := mempool.NewMockVoteMempool(ctrl)

	coinsMem := mempool.NewMockCoinsMempool(ctrl)

	actionsMem := mempool.NewMockActionMempool(ctrl)

	am := actionmanager.NewMockLastActionManager(ctrl)
	am.EXPECT().GetNonce().AnyTimes()

	ks := keystore.NewMockKeystore(ctrl)
	ks.EXPECT().OpenKeystore().Times(1)
	ks.EXPECT().GetValidatorKey(gomock.Any()).AnyTimes()

	prop, err := proposer.NewProposer(log, param, bc, host, voteMem, coinsMem, actionsMem, am, ks)
	assert.NoError(t, err)
	assert.NotNil(t, prop)

	// proposer methods
	err = prop.Start()
	assert.NoError(t, err)

	//prop.ProposeBlocks()
	//prop.Stop()

}
