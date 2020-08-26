package mempool_test

import (
	"github.com/golang/mock/gomock"
	fuzz "github.com/google/gofuzz"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNilPointer(t *testing.T) {
	// create a deposit with a nil field, test if method handles it or panics
	f := fuzz.New().NilChance(0)

	var cs primitives.CoinsState
	f.Fuzz(&cs)

	// add some balance to premine addr
	priv := testdata.PremineAddr
	pub := priv.PublicKey()
	accBytes, _ := priv.PublicKey().Hash()
	cs.Balances[accBytes] = 10000000000

	//create a deposit
	validatorPriv := validatorKeys1[0]
	validatorPub := validatorPriv.PublicKey()
	validatorPubBytes := validatorPub.Marshal()
	validatorPubHash := chainhash.HashH(validatorPubBytes[:])
	validatorProofOfPossession := validatorPriv.Sign(validatorPubHash[:])

	addr, err := pub.Hash()
	assert.NoError(t, err)
	var p [48]byte
	var s [96]byte
	copy(p[:], validatorPubBytes)
	copy(s[:], validatorProofOfPossession.Marshal())
	depositData := &primitives.DepositData{
		PublicKey:         p,
		ProofOfPossession: s,
		WithdrawalAddress: addr,
	}

	buf, err := depositData.Marshal()
	assert.NoError(t, err)

	depositHash := chainhash.HashH(buf)

	depositSig := priv.Sign(depositHash[:])

	var pubKey [48]byte
	var ds [96]byte
	copy(pubKey[:], pub.Marshal())
	copy(ds[:], depositSig.Marshal())
	// depositData will be nil
	deposit := &primitives.Deposit{
		PublicKey: pubKey,
		Signature: ds,
		//Data: depositData,
	}
	mState := state.NewState(cs, validatorsGlobal, genesisHash, param)
	assert.NotPanics(t, func() {
		_ = mState.IsDepositValid(deposit, param)
	})

}

func TestActionMempool_New(t *testing.T) {
	h, err := mockNet.GenPeer()
	assert.NoError(t, err)

	g, err := pubsub.NewGossipSub(ctx, h)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)

	host := hostnode.NewMockHostNode(ctrl)
	host.EXPECT().Topic(p2p.MsgDepositCmd).Return(g.Join(p2p.MsgVoteCmd))
	host.EXPECT().Topic(p2p.MsgExitCmd).Return(g.Join(p2p.MsgExitCmd))
	host.EXPECT().Topic(p2p.MsgGovernanceCmd).Return(g.Join(p2p.MsgGovernanceCmd))
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

	am, err := mempool.NewActionMempool(ctx, log, param, ch, host)
	assert.NoError(t, err)
	assert.NotNil(t, am)
}
