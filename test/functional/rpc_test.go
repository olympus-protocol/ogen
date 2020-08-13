// +build rpc_test

package rpc_test

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"os"
	"testing"
	"time"

	"github.com/olympus-protocol/ogen/internal/bdb"
	"github.com/olympus-protocol/ogen/internal/chainrpc"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/server"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/olympus-protocol/ogen/proto"
	testdata "github.com/olympus-protocol/ogen/test"
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

//some common vars that will be used on multiple tests
var rawTx string
var tx *primitives.Tx
var savedWallet bls.KeyPair
var ogValidators []*bls.SecretKey
var acc1 *bls.SecretKey
var acc2 *bls.SecretKey

// RPC Functional test
// 1. Start a new chain with a single node moving.
// 2. Use all the RPC methods trough a RPC Client and check all calls.
func TestMain(m *testing.M) {
	_ = os.RemoveAll(testdata.Node1Folder)
	_ = os.RemoveAll(testdata.Node2Folder)

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
	ks := keystore.NewKeystore(testdata.Node1Folder, log)

	err = ks.CreateKeystore()
	if err != nil {
		log.Fatal(err)
	}

	// Generate 128 validators
	ogValidators, err = ks.GenerateNewValidatorKey(128)
	if err != nil {
		log.Fatal(err)
	}

	err = ks.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Convert the validator to initialization params. The validator will be binded to the premineAddr
	var validators []primitives.ValidatorInitialization
	for _, vk := range ogValidators {
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
	err = S.Proposer.OpenKeystore()
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

	assert.Equal(t, strconv.Itoa(int(accBalance/1e8)), res.Balance.Confirmed)
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

	txInfo, err := C.chain.GetTransaction(ctx, &proto.Hash{Hash: txReceipt.Hash})
	assert.NoError(t, err)
	assert.NotNil(t, txInfo)
	assert.IsType(t, &proto.Tx{}, txInfo)

	assert.Equal(t, txReceipt.Hash, txInfo.Hash)
}

func Test_Validator_GetValidatorsList(t *testing.T) {
	ctx := context.Background()

	res, err := C.validators.GetValidatorsList(ctx, &proto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.ValidatorsRegistry{}, res)
	// should be same length
	valkeys, err := S.Proposer.Keystore.GetValidatorKeys()
	assert.NoError(t, err)
	assert.Equal(t, len(valkeys), len(res.Validators))
}

func Test_Validator_GetAccountValidators(t *testing.T) {
	ctx := context.Background()

	account, err := testdata.PremineAddr.PublicKey().ToAccount()
	assert.NoError(t, err)
	res, err := C.validators.GetAccountValidators(ctx, &proto.Account{Account: account})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.ValidatorsRegistry{}, res)
	// validators from premine account should be the same that ogValidators
	assert.Equal(t, len(ogValidators), len(res.Validators))

}

func Test_Network_GetNetworkInfo(t *testing.T) {
	ctx := context.Background()

	res, err := C.network.GetNetworkInfo(ctx, &proto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.NetworkInfo{}, res)
	// should be same Id
	assert.Equal(t, S.HostNode.GetHost().ID().String(), res.ID)
}

func Test_Network_GetPeersInfo(t *testing.T) {
	ctx := context.Background()

	res, err := C.network.GetPeersInfo(ctx, &proto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Peers{}, res)
	//they should have the same Peers
	assert.Equal(t, len(S.HostNode.GetPeerInfos()), len(res.Peers))
}

func Test_Network_AddPeer(t *testing.T) {
	ctx := context.Background()
	// add node
	res, err := C.network.AddPeer(ctx, &proto.IP{Host: "/ip4/127.0.0.1/tcp/24126/p2p/12D3KooWCnt52MYKVLn6fhKCoKy6HsNejEtxUt9MUwcpj1LYU2N1"})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Success{}, res)
}

func Test_Utils_StopProposer(t *testing.T) {
	ctx := context.Background()

	res, err := C.utils.StopProposer(ctx, &proto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Success{}, res)
}

func Test_Utils_StartProposer(t *testing.T) {
	ctx := context.Background()

	res, err := C.utils.StartProposer(ctx, &proto.Password{Password: testdata.KeystorePass})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Success{}, res)
}

func Test_Utils_SubmitRawData(t *testing.T) {
	ctx := context.Background()
	//create raw tx
	// random addr
	secondAccount := bls.RandKey()
	secondAddr, _ := secondAccount.PublicKey().ToAccount()
	_, data, err := bech32.Decode(secondAddr)
	if err != nil {
		t.Fatal(err)
	}

	var toPkh [20]byte
	copy(toPkh[:], data)

	var p [48]byte
	copy(p[:], testdata.PremineAddr.PublicKey().Marshal())

	tx = &primitives.Tx{
		To:            toPkh,
		FromPublicKey: p,
		Amount:        11,
		Nonce:         0,
		Fee:           1,
	}

	sigMsg := tx.SignatureMessage()
	sig := testdata.PremineAddr.Sign(sigMsg[:])
	var s [96]byte
	copy(s[:], sig.Marshal())
	tx.Signature = s

	byteTx, _ := tx.Marshal()
	rawTx = hex.EncodeToString(byteTx)
	res, err := C.utils.SubmitRawData(ctx, &proto.RawData{Data: rawTx, Type: "tx"})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Success{}, res)
	// give some time to the transaction to be added to a block
	time.Sleep(time.Second * 30)
	balance, _ := C.chain.GetAccountInfo(ctx, &proto.Account{Account: secondAddr})

	assert.Equal(t, "11", balance.Balance.Confirmed)

}

//not working
func Test_Utils_GenValidatorKey(t *testing.T) {
	/*ctx := context.Background()

	res, err := C.utils.GenValidatorKey(ctx, &proto_def.GenValidatorKeys{Keys: uint64(2), Password: testdata.KeystorePass})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto_def.KeyPairs{}, res)*/
}

func Test_Utils_DecodeRawTransaction(t *testing.T) {
	ctx := context.Background()

	res, err := C.utils.DecodeRawTransaction(ctx, &proto.RawData{Data: rawTx})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Tx{}, res)
	// compare hash, To field and FromPublicKey field
	assert.Equal(t, tx.Hash().String(), res.Hash)
	assert.Equal(t, hex.EncodeToString(tx.To[:]), res.To)
	assert.Equal(t, hex.EncodeToString(tx.FromPublicKey[:]), res.FromPublicKey)
}

func Test_Utils_DecodeRawBlock(t *testing.T) {
	ctx := context.Background()
	hash := S.Chain.State().Tip().Hash
	block, err := S.Chain.GetRawBlock(hash)
	assert.NoError(t, err)
	rawBlock := hex.EncodeToString(block)
	res, err := C.utils.DecodeRawBlock(ctx, &proto.RawData{Data: rawBlock})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Block{}, res)
	assert.Equal(t, rawBlock, res.RawBlock)
}

func Test_Wallet_CreateWallet(t *testing.T) {
	ctx := context.Background()

	res, err := C.wallet.CreateWallet(ctx, &proto.WalletReference{Name: "username", Password: "testPassword"})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.KeyPair{}, res)
	// save the returned pubkey
	savedWallet.Public = res.Public
}

func Test_Wallet_ListWallets(t *testing.T) {
	ctx := context.Background()

	res, err := C.wallet.ListWallets(ctx, &proto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Wallets{}, res)
	fmt.Println(res)
	fmt.Println(len(res.Wallets))
}

func Test_Wallet_OpenWallet(t *testing.T) {
	ctx := context.Background()

	res, err := C.wallet.OpenWallet(ctx, &proto.WalletReference{Name: "username", Password: "testPassword"})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Success{}, res)
}

func Test_Wallet_CloseWallet(t *testing.T) {
	ctx := context.Background()

	res, err := C.wallet.CloseWallet(ctx, &proto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Success{}, res)

}

func Test_Wallet_ImportWallet(t *testing.T) {
	ctx := context.Background()

	privKey, err := testdata.PremineAddr.ToWIF()
	assert.NoError(t, err)

	res, err := C.wallet.ImportWallet(ctx, &proto.ImportWalletData{Name: "premineAcc", Key: &proto.KeyPair{Private: privKey}})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.KeyPair{}, res)
	//public key must match
	acc, err := testdata.PremineAddr.PublicKey().ToAccount()
	assert.NoError(t, err)
	assert.Equal(t, acc, res.Public)
}

func Test_Wallet_DumpWallet(t *testing.T) {
	ctx := context.Background()

	privKey, err := testdata.PremineAddr.ToWIF()
	assert.NoError(t, err)
	res, err := C.wallet.DumpWallet(ctx, &proto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.KeyPair{}, res)
	fmt.Println(privKey)
	fmt.Println(res.Private)
	assert.Equal(t, privKey, res.Private)
}

func Test_Wallet_GetBalance(t *testing.T) {
	ctx := context.Background()

	res, err := C.wallet.GetBalance(ctx, &proto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Balance{}, res)

	accByte, err := testdata.PremineAddr.PublicKey().Hash()
	assert.NoError(t, err)

	balanceMap := S.Chain.State().TipState().CoinsState.Balances
	accBalance := balanceMap[accByte]

	responseBalance, err := strconv.Atoi(res.Confirmed)
	assert.NoError(t, err)


	assert.Equal(t, accBalance, uint64(responseBalance) * 1e8)
}

func Test_Wallet_GetValidators(t *testing.T) {
	ctx := context.Background()

	res, err := C.wallet.GetValidators(ctx, &proto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.ValidatorsRegistry{}, res)
	// validator length should be the same as PremineAddr (because we imported that account into wallet)
	assert.Equal(t, len(ogValidators), len(res.Validators))
}

func Test_Wallet_GetAccount(t *testing.T) {
	ctx := context.Background()

	res, err := C.wallet.GetAccount(ctx, &proto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.KeyPair{}, res)

	// Received account should be the same as PremineAddr
	acc, err := testdata.PremineAddr.PublicKey().ToAccount()
	assert.NoError(t, err)

	assert.Equal(t, acc, res.Public)
}

func Test_Wallet_SendTransaction(t *testing.T) {
	ctx := context.Background()
	secondAccount := bls.RandKey()
	secondAddr, _ := secondAccount.PublicKey().ToAccount()
	res, err := C.wallet.SendTransaction(ctx, &proto.SendTransactionInfo{Account: secondAddr, Amount: "23"})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Hash{}, res)

	time.Sleep(time.Second * 30)
	secondaccByte, err := secondAccount.PublicKey().Hash()
	assert.NoError(t, err)
	balanceMap := S.Chain.State().TipState().CoinsState.Balances
	secondaccBalance := balanceMap[secondaccByte]
	assert.Equal(t, strconv.Itoa(int(secondaccBalance/1000)), "23")
	balance, _ := C.chain.GetAccountInfo(ctx, &proto.Account{Account: secondAddr})

	assert.Equal(t, "23", balance.Balance.Confirmed)

}

func Test_Wallet_StartValidator(t *testing.T) {
	ctx := context.Background()
	secondAccount := ogValidators[0]
	privKey, err := secondAccount.ToWIF()
	assert.NoError(t, err)
	res, err := C.wallet.StartValidator(ctx, &proto.KeyPair{Private: privKey})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Success{}, res)
}

func Test_Wallet_ExitValidator(t *testing.T) {
	ctx := context.Background()
	secondAccount := ogValidators[0]
	secondAddr, _ := secondAccount.PublicKey().ToAccount()
	res, err := C.wallet.ExitValidator(ctx, &proto.KeyPair{Public: secondAddr})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Success{}, res)
}

// methods from the RPC, but not in the cli

func Test_Wallet_ValidatorBulk(t *testing.T) {
	ctx := context.Background()
	acc1 = bls.RandKey()
	secret1, err := acc1.ToWIF()
	assert.NoError(t, err)
	acc2 = bls.RandKey()
	secret2, err := acc2.ToWIF()
	assert.NoError(t, err)

	res, err := C.wallet.StartValidatorBulk(ctx, &proto.KeyPairs{Keys: []string{secret1, secret2}})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Success{}, res)
}

func Test_Wallet_ExitValidatorBulk(t *testing.T) {
	ctx := context.Background()
	pub1, err := acc1.PublicKey().ToAccount()
	assert.NoError(t, err)
	pub2, err := acc1.PublicKey().ToAccount()
	assert.NoError(t, err)
	res, err := C.wallet.ExitValidatorBulk(ctx, &proto.KeyPairs{Keys: []string{pub1, pub2}})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.IsType(t, &proto.Success{}, res)
}

func Test_Chain_Sync(t *testing.T) {
	ctx := context.Background()
	hash := S.Chain.State().Tip().Hash
	res, err := C.chain.Sync(ctx, &proto.Hash{Hash: hash.String()})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func Test_Chain_SubscribeBlocks(t *testing.T) {
	ctx := context.Background()
	res, err := C.chain.SubscribeBlocks(ctx, &proto.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func Test_Chain_SubscribeTransactions(t *testing.T) {
	ctx := context.Background()
	pub1, err := acc1.PublicKey().ToAccount()
	assert.NoError(t, err)
	pub2, err := acc1.PublicKey().ToAccount()
	assert.NoError(t, err)
	res, err := C.chain.SubscribeTransactions(ctx, &proto.KeyPairs{Keys: []string{pub1, pub2}})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func Test_Chain_SubscribeValidatorTransactions(t *testing.T) {
	ctx := context.Background()
	//validator pubkeys
	pub1, err := ogValidators[0].PublicKey().ToAccount()
	assert.NoError(t, err)
	pub2, err := ogValidators[1].PublicKey().ToAccount()
	assert.NoError(t, err)
	res, err := C.chain.SubscribeValidatorTransactions(ctx, &proto.KeyPairs{Keys: []string{pub1, pub2}})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
