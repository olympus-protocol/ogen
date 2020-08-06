// +build rpc_test

package rpc_test

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/bls"
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
	"strconv"
)

type RPC struct {
	conn       *grpc.ClientConn
	chain      proto.ChainClient
	validators proto.ValidatorsClient
	utils      proto.UtilsClient
	network    proto.NetworkClient
	wallet     proto.WalletClient
}

var C *RPC
var S *server.Server
var SAddr peer.AddrInfo

var B *server.Server

var initParams primitives.InitializationParameters

// RPC Functional test
// 1. Start a new chain with a single node moving.
// 2. Use all the RPC methods trough a RPC Client and check all calls.
func TestMain(m *testing.M) {
	// Create the node server and the clients for global usage
	startNode()

	// Start secondary node
	secondNode()

	// Run the test functions.
	os.Exit(m.Run())
}

func startNode() {

	// Create datafolder
	err := os.Mkdir(testdata.Node1Folder, 0777)

	// Initialize the logger
	log := logger.New(os.Stdin)
	log.WithDebug()

	// Create the premine address on bech32 format.
	addr, err := testdata.PremineAddr.PublicKey().ToAccount()
	if err != nil {
		log.Fatal(err)
	}

	// Create a keystore
	log.Info("Creating keystore")
	ks, err := keystore.NewKeystore(testdata.Node1Folder, log, testdata.KeystorePass)
	if err != nil {
		log.Fatal(err)
	}

	// Generate 128 validators
	valData, err := ks.GenerateNewValidatorKey(128, testdata.KeystorePass)
	if err != nil {
		log.Fatal(err)
	}

	err = ks.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Convert the validator to initialization params.
	var validators []primitives.ValidatorInitialization
	for _, vk := range valData {
		val := primitives.ValidatorInitialization{
			PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
			PayeeAddress: addr,
		}
		validators = append(validators, val)
	}

	// Load the block database
	db, err := bdb.NewBlockDB(testdata.Node1Folder, testdata.IntTestParams, log)
	if err != nil {
		log.Fatal(err)
	}

	// Create the server instance

	// Get the configuration params from the testdata
	c := testdata.Conf
	c.LogFile = true
	c.RPCWallet = true

	// Override the data folder.
	c.DataFolder = testdata.Node1Folder

	// Create the server instance.
	ctx := context.Background()

	// Create the initialization parameters
	initParams = primitives.InitializationParameters{
		GenesisTime:       time.Unix(time.Now().Unix()+15, 0),
		PremineAddress:    addr,
		InitialValidators: validators,
	}

	S, err = server.NewServer(ctx, &c, log, testdata.IntTestParams, db, initParams)
	if err != nil {
		log.Fatal(err)
	}

	SAddr = peer.AddrInfo{
		ID:    S.HostNode.GetHost().ID(),
		Addrs: S.HostNode.GetHost().Network().ListenAddresses(),
	}

	// Start the server
	go S.Start()

	// Initialize the RPC Client
	err = rpcClient()
	if err != nil {
		log.Fatal(err)
	}

	// Open the Keystore to start generating blocks
	err = S.Proposer.OpenKeystore(testdata.KeystorePass)
	err = S.Proposer.Start()

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
	conn, err := grpc.Dial("127.0.0.1:"+testdata.Conf.RPCPort, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}

	// Initialize the clients from the protobuf
	C = &RPC{
		conn:       conn,
		chain:      proto.NewChainClient(conn),
		validators: proto.NewValidatorsClient(conn),
		utils:      proto.NewUtilsClient(conn),
		network:    proto.NewNetworkClient(conn),
		wallet:     proto.NewWalletClient(conn),
	}
	return nil
}

func secondNode() {
	// Create datafolder
	_ = os.Mkdir(testdata.Node2Folder, 0777)

	// Create logger
	logfile, err := os.Create(testdata.Node2Folder + "/log.log")
	if err != nil {
		panic(err)
	}
	log := logger.New(logfile)
	log.WithDebug()

	// Load the block database
	db, err := bdb.NewBlockDB(testdata.Node2Folder, testdata.IntTestParams, log)
	if err != nil {
		log.Fatal(err)
	}

	// Load the conf params
	c := testdata.Conf

	// Override the datafolder.
	c.DataFolder = testdata.Node2Folder
	c.RPCPort = "25001"
	c.Port = "24131"

	// Create the server instance
	B, err = server.NewServer(context.Background(), &c, log, testdata.IntTestParams, db, initParams)
	if err != nil {
		log.Fatal(err)
	}
	// Start the server
	go B.Start()
}

func Test_Connections(t *testing.T) {
	// The backup node should connect to the first node
	err := B.HostNode.GetHost().Connect(context.TODO(), SAddr)
	assert.NoError(t, err)
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

func Test_Chain_GetRawBlock(t *testing.T) {
	ctx := context.Background()
	hash := S.Chain.State().Tip().Hash
	res, err := C.chain.GetRawBlock(ctx, &proto.Hash{Hash: hash.String()})

	assert.NoError(t, err)
	assert.NotNil(t, res)

	assert.IsType(t, &proto.Block{}, res)

	block, err := S.Chain.GetRawBlock(hash)
	assert.NoError(t, err)

	assert.Equal(t, hex.EncodeToString(block), res.RawBlock)
}

func Test_Chain_GetBlockHash(t *testing.T) {
	ctx := context.Background()

	tip := S.Chain.State().Tip()

	res, err := C.chain.GetBlockHash(ctx, &proto.Number{Number: uint64(tip.Height)})

	assert.NoError(t, err)
	assert.NotNil(t, res)

	assert.IsType(t, &proto.Hash{}, res)

	assert.Equal(t, tip.Hash.String(), res.Hash)
}

func Test_Chain_GetBlock(t *testing.T) {
	ctx := context.Background()
	hash := S.Chain.State().Tip().Hash
	res, err := C.chain.GetBlock(ctx, &proto.Hash{Hash: hash.String()})

	assert.NoError(t, err)
	assert.NotNil(t, res)

	assert.IsType(t, &proto.Block{}, res)

	block, err := S.Chain.GetBlock(hash)
	assert.NoError(t, err)

	assert.Equal(t, block.Hash().String(), res.Hash)
}

func Test_Chain_GetAccountInfo(t *testing.T) {
	ctx := context.Background()

	account, err := testdata.PremineAddr.PublicKey().ToAccount()
	assert.NoError(t, err)

	res, err := C.chain.GetAccountInfo(ctx, &proto.Account{Account: account})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.AccountInfo{}, res)

	accByte, err := testdata.PremineAddr.PublicKey().Hash()
	assert.NoError(t, err)
	balanceMap := S.Chain.State().TipState().CoinsState.Balances
	accBalance := balanceMap[accByte]

	assert.Equal(t, strconv.Itoa(int(accBalance/1000)), res.Balance.Confirmed)
}

func Test_Chain_GetTransaction(t *testing.T) {
	ctx := context.Background()

	//make a transaction using premine account
	privKey, _ := testdata.PremineAddr.ToWIF()

	response, err := C.wallet.ImportWallet(ctx, &proto.ImportWalletData{Name: "premineAddrAccount", Key: &proto.KeyPair{Private: privKey}})
	assert.Nil(t, err)
	assert.NotNil(t, response)
	randomAddr, _ := bls.RandKey().PublicKey().ToAccount()
	txReceipt, err := C.wallet.SendTransaction(ctx, &proto.SendTransactionInfo{Account: randomAddr, Amount: "10"})
	assert.Nil(t, err)
	assert.NotNil(t, txReceipt)
	// give some time to the transaction to be added to a block
	time.Sleep(time.Second * 20)
	fmt.Println(txReceipt)

	txInfo, err := C.chain.GetTransaction(ctx, &proto.Hash{Hash: txReceipt.Hash})
	assert.NoError(t, err)
	assert.NotNil(t, txInfo)
	assert.IsType(t, &proto.Tx{}, txInfo)

	assert.Equal(t, txReceipt.Hash, txInfo.Hash)
}
