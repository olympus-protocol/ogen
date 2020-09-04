package state_test

import (
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

var secrets = make([]*bls.SecretKey, 50)
var publics = make([]*bls.PublicKey, 50)
var validators = make([]*primitives.Validator, 50)
var params = &testdata.TestParams
var owner = bls.RandKey()

func init() {
	for i := range secrets {
		key := bls.RandKey()
		secrets[i] = key
		publics[i] = key.PublicKey()
		var pub [48]byte
		copy(pub[:], key.PublicKey().Marshal())
		payee, _ := owner.PublicKey().Hash()
		validators[i] = &primitives.Validator{
			Balance:          100 * params.UnitsPerCoin,
			PubKey:           pub,
			PayeeAddress:     payee,
			Status:           primitives.StatusActive,
			FirstActiveEpoch: 0,
			LastActiveEpoch:  0,
		}
	}
}

func TestState(t *testing.T) {
	cs := primitives.CoinsState{
		Balances: make(map[[20]byte]uint64),
		Nonces:   make(map[[20]byte]uint64),
	}

	gs := primitives.Governance{
		ReplaceVotes:   make(map[[20]byte]chainhash.Hash),
		CommunityVotes: make(map[chainhash.Hash]primitives.CommunityVoteData),
	}

	gen := primitives.GetGenesisBlock()

	initState := state.NewState(cs, gs, validators, gen.Hash(), params)

	assert.Equal(t, validators, initState.GetValidators().Validators)

	rawState, err := initState.Marshal()
	assert.NoError(t, err)

	s := state.NewEmptyState()
	err = s.Unmarshal(rawState)
	assert.NoError(t, err)

	assert.Equal(t, initState, s)
}
