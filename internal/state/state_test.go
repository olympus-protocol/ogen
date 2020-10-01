package state_test

import (
	"encoding/hex"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var NumValidators = 10

type pair struct {
	public  *bls.PublicKey
	private *bls.SecretKey
}

var secrets = make([]pair, NumValidators)

var params = testdata.TestParams
var premineAddr = bls.RandKey()

var initParams *initialization.InitializationParameters

var blocksMap = make(map[uint64]chainhash.Hash)

func init() {

	initParams = &initialization.InitializationParameters{
		InitialValidators: make([]initialization.ValidatorInitialization, NumValidators),
		PremineAddress:    premineAddr.PublicKey().ToAccount(),
		GenesisTime:       time.Now(),
	}

	for i := range secrets {
		key := bls.RandKey()
		secrets[i] = pair{
			public:  key.PublicKey(),
			private: key,
		}
		var pub [48]byte
		copy(pub[:], key.PublicKey().Marshal())
		initParams.InitialValidators[i] = initialization.ValidatorInitialization{
			PubKey:       hex.EncodeToString(key.PublicKey().Marshal()),
			PayeeAddress: premineAddr.PublicKey().ToAccount(),
		}
	}
}

func TestState(t *testing.T) {
	gen := primitives.GetGenesisBlock()
	blocksMap[0] = gen.Hash()

	initState, err := state.GetGenesisStateWithInitializationParameters(gen.Hash(), initParams, &params)
	assert.NoError(t, err)

	ser, err := initState.Marshal()
	assert.NoError(t, err)

	s := state.NewEmptyState()
	err = s.Unmarshal(ser)

	assert.NoError(t, err)
}
