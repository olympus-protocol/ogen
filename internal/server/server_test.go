package server_test

import (
	"context"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/server"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

// ctx is the global context used for the entire test
var ctx = context.Background()

// params are the params used on the test
var param = &testdata.TestParams

// validatorKeys is a slice of signatures that match the validators index
var validatorKeys1 []*bls.SecretKey
var validatorKeys2 []*bls.SecretKey

// validators are the initial validators on the realState
var validators1 []*primitives.Validator
var validators2 []*primitives.Validator

// init params used on the test
var stateParams state.InitializationParameters

func init() {
	priv := testdata.PremineAddr

	addrByte, _ := priv.PublicKey().Hash()
	addr := testdata.PremineAddr.PublicKey().ToAccount()

	for i := 0; i < 100; i++ {
		if i < 50 {
			key := bls.RandKey()
			validatorKeys1 = append(validatorKeys1, bls.RandKey())
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
			key := bls.RandKey()
			validatorKeys2 = append(validatorKeys2, bls.RandKey())
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

func TestServer_Object(t *testing.T) {

	log := logger.New(os.Stdin)

	db, err := blockdb.NewMemoryDB(testdata.TestParams, log)
	assert.NoError(t, err)

	serv, err := server.NewServer(ctx, &testdata.Conf, log, param, db, stateParams)
	assert.NoError(t, err)
	assert.NotNil(t, serv)

	cleanTestData()

}

func cleanTestData() {
	_ = os.RemoveAll("cert")
	_ = os.RemoveAll("peerstore")
	_ = os.Remove("./net.db")
	_ = os.Remove("./keystore.db")
	_ = os.Remove("./node_key.dat")
}
