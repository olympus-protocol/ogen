// +build send_test

package send_test

// TODO Rebuild test
//
//import (
//	"context"
//	"crypto/tls"
//	"encoding/hex"
//	"fmt"
//	"github.com/libp2p/go-libp2p-core/peer"
//	"github.com/olympus-protocol/ogen/api/proto"
//	"github.com/olympus-protocol/ogen/internal/blockdb"
//	"github.com/olympus-protocol/ogen/internal/chainrpc"
//	"github.com/olympus-protocol/ogen/internal/config"
//	"github.com/olympus-protocol/ogen/internal/keystore"
//	"github.com/olympus-protocol/ogen/internal/logger"
//	"github.com/olympus-protocol/ogen/internal/server"
//	"github.com/olympus-protocol/ogen/internal/state"
//	"github.com/olympus-protocol/ogen/pkg/bech32"
//	"github.com/olympus-protocol/ogen/pkg/bls"
//	"github.com/olympus-protocol/ogen/pkg/primitives"
//	testdata "github.com/olympus-protocol/ogen/test"
//	"github.com/stretchr/testify/assert"
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/credentials"
//	"os"
//	"strconv"
//	"testing"
//	"time"
//)
//
//type RPC struct {
//	conn       *grpc.ClientConn
//	chain      proto.ChainClient
//	validators proto.ValidatorsClient
//	utils      proto.UtilsClient
//	network    proto.NetworkClient
//	wallet     proto.WalletClient
//}
//
//var server1 *server.Server
//var server2 *server.Server
//var hostMultiAddr peer.AddrInfo
//var C *RPC
//
//// Send test.
//// 1. The initial node will be created with a premine address
//// 2. The second node will connect to the intial node
//// 3. Two transaction will be broadcasted with the same nonce, one after another. The first one is expected to be
////	  accepted. The second one should be stopped and never be added to mempool.
//
//func TestMain(m *testing.M) {
//
//	//modify testdata for this test
//	testdata.Conf.RPCWallet = true
//	testdata.Conf.LogFile = true
//
//	// Create datafolder
//	_ = os.Mkdir(testdata.Node1Folder, 0777)
//
//	// Create logger
//	logfile, err := os.Create(testdata.Node1Folder + "/log.log")
//	if err != nil {
//		panic(err)
//	}
//	log := logger.New(logfile)
//	log.WithDebug()
//
//	// Create a keystore
//	log.Info("Creating keystore")
//	ks := keystore.NewKeystore(testdata.Node1Folder, log)
//
//	err = ks.CreateKeystore()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	validatorKeys, err := ks.GenerateNewValidatorKey(128)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Infof("Generated %v keys", len(validatorKeys))
//	err = ks.Close()
//	if err != nil {
//		log.Fatal(err)
//	}
//	addr, err := testdata.PremineAddr.PublicKey().ToAccount()
//	if err != nil {
//		log.Fatal(err)
//	}
//	validators := []state.ValidatorInitialization{}
//	for _, vk := range validatorKeys {
//		val := state.ValidatorInitialization{
//			PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
//			PayeeAddress: addr,
//		}
//		validators = append(validators, val)
//	}
//
//	// Create the initialization parameters
//	ip := state.InitializationParameters{
//		GenesisTime:       time.Now(),
//		PremineAddress:    addr,
//		InitialValidators: validators,
//	}
//	// Load the block database
//	db, err := blockdb.NewBlockDB(testdata.Node1Folder, testdata.IntTestParams, log)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create the server instance
//	ctx, cancel := context.WithCancel(context.Background())
//	config.InterruptListener(log, cancel)
//	c := testdata.Conf
//	c.DataFolder = testdata.Node1Folder
//	server1, err = server.NewServer(ctx, &c, log, testdata.IntTestParams, db, ip)
//	if err != nil {
//		log.Fatal(err)
//	}
//	go server1.Start()
//	//os.Exit(m.Run())
//	var initialValidators []state.ValidatorInitialization
//	for _, sv := range server1.Chain.State().TipState().ValidatorRegistry {
//		initialValidators = append(initialValidators, state.ValidatorInitialization{
//			PubKey:       hex.EncodeToString(sv.PubKey[:]),
//			PayeeAddress: bech32.Encode(testdata.IntTestParams.AccountPrefixes.Public, sv.PayeeAddress[:]),
//		})
//	}
//	ip.InitialValidators = initialValidators
//	hostMultiAddr.Addrs = server1.HostNode.GetHost().Addrs()
//	hostMultiAddr.ID = server1.HostNode.GetHost().ID()
//	// Open the Keystore to start generating blocks
//	err = server1.Proposer.OpenKeystore()
//	if err != nil {
//		log.Fatal(err)
//	}
//	err = server1.Proposer.Start()
//	if err != nil {
//		log.Fatal(err)
//	}
//	err = rpcClient()
//	go runSecondNode(server1, ip, m)
//
//	<-ctx.Done()
//	db.Close()
//	err = server1.Stop()
//	if err != nil {
//		log.Fatal(err)
//	}
//}
//
//func runSecondNode(ps *server.Server, ip state.InitializationParameters, m *testing.M) {
//	_ = os.Mkdir(testdata.Node2Folder, 0777)
//	logfile, err := os.Create(testdata.Node2Folder + "/log.log")
//	if err != nil {
//		panic(err)
//	}
//	log := logger.New(logfile)
//	log.WithDebug()
//	ctx, cancel := context.WithCancel(context.Background())
//	config.InterruptListener(log, cancel)
//
//	// Create a keystore
//	log.Info("Creating keystore")
//	ks := keystore.NewKeystore(testdata.Node2Folder, log)
//	if err != nil {
//		log.Fatal(err)
//	}
//	validatorKeys, err := ks.GenerateNewValidatorKey(128)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Infof("Generated %v keys", len(validatorKeys))
//	err = ks.Close()
//	if err != nil {
//		log.Fatal(err)
//	}
//	db, err := blockdb.NewBlockDB(testdata.Node2Folder, testdata.IntTestParams, log)
//	if err != nil {
//		log.Fatal(err)
//	}
//	secondNodeConf := &config.Config{
//		DataFolder:   testdata.Node2Folder,
//		NetworkName:  "integration tests net",
//		InitialNodes:     []peer.AddrInfo{hostMultiAddr},
//		Port:         "24000",
//		RPCProxy:     false,
//		RPCProxyPort: "8080",
//		RPCPort:      "24001",
//		RPCWallet:    false,
//		Debug:        false,
//		LogFile:      false,
//		Pprof:        false,
//	}
//	server2, err = server.NewServer(ctx, secondNodeConf, log, testdata.IntTestParams, db, ip)
//	if err != nil {
//		log.Fatal(err)
//	}
//	go server2.Start()
//	err = server2.Proposer.OpenKeystore()
//	err = server2.Proposer.Start()
//	fmt.Println("Second node ready")
//	os.Exit(m.Run())
//	<-ctx.Done()
//	blockdb.Close()
//	err = server2.Stop()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//}
//
//func rpcClient() error {
//	// Load the certificates (this is optional).
//	certPool, err := chainrpc.LoadCerts(testdata.Node1Folder)
//	if err != nil {
//		return nil
//	}
//
//	// Create the credentials for the RPC connection.
//	creds := credentials.NewTLS(&tls.Config{
//		InsecureSkipVerify: false,
//		RootCAs:            certPool,
//	})
//	// Start the gRPC dial
//	conn, err := grpc.Dial("127.0.0.1:"+testdata.Conf.RPCPort, grpc.WithTransportCredentials(creds))
//	if err != nil {
//		return err
//	}
//
//	// Initialize the clients from the protobuf
//	C = &RPC{
//		conn:       conn,
//		chain:      proto.NewChainClient(conn),
//		validators: proto.NewValidatorsClient(conn),
//		utils:      proto.NewUtilsClient(conn),
//		network:    proto.NewNetworkClient(conn),
//		wallet:     proto.NewWalletClient(conn),
//	}
//	return nil
//}
//
//func Test_PremineAccount(t *testing.T) {
//	ctx := context.Background()
//	// Test PremineAddr in State
//
//	accByte, err := testdata.PremineAddr.PublicKey().Hash()
//	assert.Nil(t, err)
//
//	balanceMap := server1.Chain.State().TipState().CoinsState.Balances
//	premineAccBalance := balanceMap[accByte]
//
//	//Test PremineAddr balance and add premineAddr account to RPCWallet"
//
//	privKey, _ := testdata.PremineAddr.ToWIF()
//
//	response, err := C.wallet.ImportWallet(ctx, &proto.ImportWalletData{Name: "premineAddrAccount", Key: &proto.KeyPair{Private: privKey}})
//	assert.Nil(t, err)
//	assert.NotNil(t, response)
//
//	balance, err := C.wallet.GetBalance(ctx, &proto.Empty{})
//
//	assert.Equal(t, balance.Confirmed, strconv.Itoa(int(premineAccBalance/1000)))
//}
//
//func Test_SendTxs(t *testing.T) {
//	/* register each node to one another. This happens in addr/getaddr, but we want to guarantee that both nodes know
//	   each other before the test */
//	server1Addr := server1.HostNode.GetHost().Peerstore().PeerInfo(server1.HostNode.GetHost().ID())
//	server2Addr := server2.HostNode.GetHost().Peerstore().PeerInfo(server2.HostNode.GetHost().ID())
//	server1Multi, _ := peer.AddrInfoToP2pAddrs(&server1Addr)
//	server2Multi, _ := peer.AddrInfoToP2pAddrs(&server2Addr)
//	err := server1.HostNode.SavePeer(server2Multi[0])
//	assert.Nil(t, err)
//	err = server2.HostNode.SavePeer(server1Multi[0])
//	assert.Nil(t, err)
//
//	//Create an empty address to send coins
//	ctx := context.Background()
//	secondAccount := bls.RandKey()
//	secondAddr, _ := secondAccount.PublicKey().ToAccount()
//
//	//Send a first, valid transaction
//	_, err = C.wallet.SendTransaction(ctx, &proto.SendTransactionInfo{Account: secondAddr, Amount: "10"})
//	assert.Nil(t, err)
//
//	// give some time to the transaction to be added to a block
//	time.Sleep(time.Second * 30)
//
//	balance, _ := C.chain.GetAccountInfo(ctx, &proto.Account{Account: secondAddr})
//
//	assert.Equal(t, "10", balance.Balance.Confirmed)
//
//	// check if transaction is in the second node
//	pub := testdata.PremineAddr.PublicKey()
//	acc1Byte, err := pub.Hash()
//	if err != nil {
//		t.Fatal(err)
//	}
//	acc2Byte, err := secondAccount.PublicKey().Hash()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	balanceMap := server2.Chain.State().TipState().CoinsState.Balances
//	balance2 := balanceMap[acc2Byte]
//
//	assert.Equal(t, balance.Balance.Confirmed, strconv.Itoa(int(balance2/1000)))
//
//	//Create a second transaction. This time, it will be raw, and will have an incorrect nonce
//	_, data, err := bech32.Decode(secondAddr)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if len(data) != 20 {
//		t.Fatal("invalid address")
//	}
//	var toPkh [20]byte
//	copy(toPkh[:], data)
//
//	var p [48]byte
//	copy(p[:], pub.Marshal())
//
//	nonceMap := server1.Chain.State().TipState().CoinsState.Nonces
//	nonce := nonceMap[acc1Byte]
//	assert.NotNil(t, nonce)
//
//	tx := &primitives.Tx{
//		To:            toPkh,
//		FromPublicKey: p,
//		Amount:        11,
//		Nonce:         nonce,
//		Fee:           1,
//	}
//
//	sigMsg := tx.SignatureMessage()
//	sig := testdata.PremineAddr.Sign(sigMsg[:])
//	var s [96]byte
//	copy(s[:], sig.Marshal())
//	tx.Signature = s
//
//	byteTx, _ := tx.Marshal()
//	rawTx := hex.EncodeToString(byteTx)
//
//	//broadcast evil rawtx
//	msg2, err := C.utils.SubmitRawData(ctx, &proto.RawData{Data: rawTx, Type: "tx"})
//	assert.Nil(t, err)
//
//	time.Sleep(time.Second * 30)
//
//	//check that tx was not added to the blockchain
//	res, err := C.chain.GetTransaction(ctx, &proto.Hash{Hash: msg2.Data})
//	assert.NotNil(t, err)
//	assert.Nil(t, res)
//
//	// To check the banscore, comment this lines and check in the logs of the nodes
//	os.RemoveAll(testdata.Node1Folder)
//	os.RemoveAll(testdata.Node2Folder)
//
//}
