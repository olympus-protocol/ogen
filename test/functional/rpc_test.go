//+build rpc_test

package rpc_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/cli/rpcclient"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/proto"
	"github.com/olympus-protocol/ogen/server"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/olympus-protocol/ogen/utils/logger"
	"github.com/stretchr/testify/assert"
)

var C *rpcclient.RPCClient
var S *server.Server

// RPC Functional test
// 1. Start a new chain with a single node moving.
// 2. Use all the RPC methods trough a RPC Client and check all calls.
func TestMain(m *testing.M) {
	startNode(m)
}

func startNode(m *testing.M) {

	// Create datafolder
	os.Mkdir(testdata.Node1Folder, 0777)

	// Initialize the logger
	log := logger.New(os.Stdin)
	log.WithDebug()

	// Create the premine address on bech32 format.
	addr, err := testdata.PremineAddr.PublicKey().ToAddress(testdata.IntTestParams.AddrPrefix.Public)
	if err != nil {
		log.Fatal(err)
	}

	// Create a keystore
	log.Info("Creating keystore")
	keystore, err := keystore.NewKeystore(testdata.Node1Folder, log, testdata.KeystorePass)
	if err != nil {
		log.Fatal(err)
	}

	// Generate 128 validators
	valData, err := keystore.GenerateNewValidatorKey(128, testdata.KeystorePass)
	if err != nil {
		log.Fatal(err)
	}

	// Conver the validator to initialization params.
	validators := []primitives.ValidatorInitialization{}
	for _, vk := range valData {
		val := primitives.ValidatorInitialization{
			PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
			PayeeAddress: addr,
		}
		validators = append(validators, val)
	}

	// Create the initialization parameters
	ip := primitives.InitializationParameters{
		GenesisTime:       time.Unix(time.Now().Unix()+10, 0),
		PremineAddress:    addr,
		InitialValidators: validators,
	}

	// Load the block database
	bdb, err := bdb.NewBlockDB(testdata.Node1Folder, testdata.IntTestParams, log)
	if err != nil {
		log.Fatal(err)
	}

	// Create the server instance
	
	// Get the configuration params from the testdata
	c := testdata.Conf

	// Override the data folder.
	c.DataFolder = testdata.Node1Folder

	// Create the server instance.
	S, err = server.NewServer(context.Background(), &c, log, testdata.IntTestParams, bdb, ip)
	if err != nil {
		log.Fatal(err)
	}

	// Start the server
	go S.Start()

	// Open the Keystore to start generating blocks
	S.Proposer.OpenKeystore(testdata.KeystorePass)
	S.Proposer.Start()

	// Initialize the RPC Client
	rpcClient()

	// Wait 5 seconds to generate some blocks
	time.Sleep(time.Second * 5)

	// Run the test functions.
	os.Exit(m.Run())
}

func rpcClient() {
	C = rpcclient.NewRPCClient("127.0.0.1:" + testdata.Conf.RPCPort, testdata.Node1Folder)
}

func Test_Chain_GetChainInfo(t *testing.T) {

	s, err := C.GetChainInfo()

	assert.NoError(t, err)
	assert.NotNil(t, s)

	var ChainInfo proto.ChainInfo

	err = json.Unmarshal([]byte(s), &ChainInfo)

	assert.NoError(t, err)
}