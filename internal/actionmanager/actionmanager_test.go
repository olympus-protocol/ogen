package actionmanager_test

import (
	"context"
	"github.com/golang/mock/gomock"
	fuzz "github.com/google/gofuzz"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

// ctx is the global context used for the entire test
var ctx = context.Background()

// mockNet is a mock network used for PubSubs from libp2p
var mockNet = mocknet.New(ctx)

// genesisHash is just a random hash to set as genesis hash.
var genesisHash chainhash.Hash

// params are the params used on the test
var param = &testdata.TestParams

// validatorKeys is a slice of signatures that match the validators index
var validatorKeys1 []bls_interface.SecretKey
var validatorKeys2 []bls_interface.SecretKey

// validators are the initial validators on the realState
var validators1 []*primitives.Validator
var validators2 []*primitives.Validator
var validatorsGlobal []*primitives.Validator

func init() {
	f := fuzz.New().NilChance(0)
	f.Fuzz(&genesisHash)

	for i := 0; i < 100; i++ {
		if i < 50 {
			key := bls.CurrImplementation.RandKey()
			validatorKeys1 = append(validatorKeys1, bls.CurrImplementation.RandKey())
			val := &primitives.Validator{
				Balance:          100 * 1e8,
				PayeeAddress:     [20]byte{},
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
				PayeeAddress:     [20]byte{},
				Status:           primitives.StatusActive,
				FirstActiveEpoch: 0,
				LastActiveEpoch:  0,
			}
			copy(val.PubKey[:], key.PublicKey().Marshal())
			validators2 = append(validators2, val)
		}

	}
	validatorsGlobal = append(validators1, validators2...)
}

func TestLastActionManager_Instance(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Info(gomock.Any()).AnyTimes()

	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	g, err := pubsub.NewGossipSub(ctx, h)
	assert.NoError(t, err)

	s := state.NewMockState(ctrl)
	s.EXPECT().GetValidatorRegistry().AnyTimes().Return(validatorsGlobal)

	stateService := chain.NewMockStateService(ctrl)
	stateService.EXPECT().TipStateAtSlot(uint64(2)).Times(2).Return(s, nil)
	stateService.EXPECT().TipStateAtSlot(uint64(3)).Times(2).Return(s, nil)
	stateService.EXPECT().Add(gomock.Any()).AnyTimes()

	bc := chain.NewMockBlockchain(ctrl)
	bc.EXPECT().Notify(gomock.Any())
	bc.EXPECT().State().AnyTimes().Return(stateService)

	host := hostnode.NewMockHostNode(ctrl)
	host.EXPECT().Topic(p2p.MsgValidatorStartCmd).Return(g.Join(p2p.MsgVoteCmd))
	host.EXPECT().GetHost().Return(h)

	am, err := actionmanager.NewLastActionManager(ctx, host, log, bc, param)
	assert.NoError(t, err)
	assert.NotNil(t, am)

	nonce := am.GetNonce()
	assert.NotNil(t, nonce)

}
