//+build rpc_test

package rpc_test

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/chainrpc"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/proto"
	"github.com/olympus-protocol/ogen/server"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/olympus-protocol/ogen/utils/logger"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type RPC struct {
	conn    *grpc.ClientConn
	chain      proto.ChainClient
	validators proto.ValidatorsClient
	utils      proto.UtilsClient
	network    proto.NetworkClient
	wallet     proto.WalletClient
}
var C *RPC
var S *server.Server

// RPC Functional test
// 1. Start a new chain with a single node moving.
// 2. Use all the RPC methods trough a RPC Client and check all calls.
func TestMain(m *testing.M) {
	// Create the node server and the clients for global usage
	startNode()

	// Run the test functions.
	os.Exit(m.Run())
}

func startNode() {

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
		GenesisTime:       time.Unix(time.Now().Unix()+7, 0),
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
	err = rpcClient()
	if err != nil {
		log.Fatal(err)
	}

	// Wait 5 seconds to generate some blocks
	time.Sleep(time.Second * 5)
}

func rpcClient() error {
	// Load the certificates (this is optional).
	certPool, err := chainrpc.LoadCerts(testdata.Node1Folder)
	if err != nil {
		return nil
	}

	// Create the credentials for the RPC connection.
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            certPool,
	})
	// Start the gRPC dial
	conn, err := grpc.Dial("127.0.0.1:" + testdata.Conf.RPCPort, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	
	// Initialize the clients from the protobuf
	C = &RPC{
		conn: conn,
		chain:      proto.NewChainClient(conn),
		validators: proto.NewValidatorsClient(conn),
		utils:      proto.NewUtilsClient(conn),
		network:    proto.NewNetworkClient(conn),
		wallet:     proto.NewWalletClient(conn),
	}
	return nil
}

func Test_Chain_GetChainInfo(t *testing.T) {
	ctx := context.Background()
	res, err := C.chain.GetChainInfo(ctx, &proto.Empty{})

	assert.NoError(t, err)
	assert.NotNil(t, res)

	assert.IsType(t, &proto.ChainInfo{}, res)

	assert.Equal(t, S.Chain.State().Tip().Hash.String(), res.BlockHash)
	assert.Equal(t, S.Chain.State().Tip().Height, res.BlockHeight)
}