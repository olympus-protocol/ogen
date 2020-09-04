package mempool_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
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

	gs := primitives.Governance{
		ReplaceVotes:   make(map[[20]byte]chainhash.Hash),
		CommunityVotes: make(map[chainhash.Hash]primitives.CommunityVoteData),
	}

	mState := state.NewState(cs, gs, validatorsGlobal, genesisHash, param)
	assert.NotPanics(t, func() {
		_ = mState.IsDepositValid(deposit, param)
	})

}
