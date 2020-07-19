//+build sync_test

package sync_test

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/server"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/olympus-protocol/ogen/utils/logger"
)

var hostMultiAddr peer.AddrInfo

var initializationParams primitives.InitializationParameters

// Sync test.
// 1. The initial node will load pre-built chain containing 1000 blocks at a certain genesis time.
// 2. The second node will connect to the intial node and sync those 1000 blocks.
// 3. A third process will check the second node with timeouts to see if the sync is stalled at a certain point of the sync.
// 	  if the second node finishes the sync of the 1000 blocks the test pass.
func TestMain(m *testing.M) {
	// Create datafolder
	os.Mkdir(testdata.Node1Folder, 0777)

	// Create logger
	log := logger.New(os.Stdin)
	log.WithDebug()

	// Copy files from data folder
	ks, err := ioutil.ReadFile("./test_data/valid_chain_1000_blocks_gt_1595126983/keystore.db")
	if err != nil {
		log.Fatal(err)
	}
	bf, err := ioutil.ReadFile("./test_data/valid_chain_1000_blocks_gt_1595126983/chain.db")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(path.Join(testdata.Node1Folder, "keystore.db"), ks, 0777)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(path.Join(testdata.Node1Folder, "chain.db"), bf, 0777)
	if err != nil {
		log.Fatal(err)
	}
	// Create a keystore
	log.Info("Opening keystore")
	keystore, err := keystore.NewKeystore(testdata.Node1Folder, log, testdata.KeystorePass)
	if err != nil {
		log.Fatal(err)
	}
	validatorKeys, err := keystore.GetValidatorKeys()
	if err != nil {
		log.Fatal(err)
	}
	keystore.Close()
	addr, err := testdata.PremineAddr.PublicKey().ToAddress(testdata.IntTestParams.AddrPrefix.Public)
	if err != nil {
		log.Fatal(err)
	}
	validators := []primitives.ValidatorInitialization{}
	for _, vk := range validatorKeys {
		val := primitives.ValidatorInitialization{
			PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
			PayeeAddress: addr,
		}
		validators = append(validators, val)
	}
	
	// Create the initialization parameters
	initializationParams = primitives.InitializationParameters{
		GenesisTime:       time.Unix(1595126983, 0),
		PremineAddress:    addr,
		InitialValidators: validators,
	}
	// Load the block database
	bdb, err := bdb.NewBlockDB(testdata.Node1Folder, testdata.IntTestParams, log)
	if err != nil {
		log.Fatal(err)
	}

	// Create the server instance
	ctx, cancel := context.WithCancel(context.Background())
	config.InterruptListener(log, cancel)
	c := testdata.Conf
	c.DataFolder = testdata.Node1Folder
	s, err := server.NewServer(ctx, &c, log, testdata.IntTestParams, bdb, initializationParams)
	if err != nil {
		log.Fatal(err)
	}

	hostMultiAddr.Addrs = s.HostNode.GetHost().Addrs()
	hostMultiAddr.ID = s.HostNode.GetHost().ID()

	go s.Start()
	go runSecondNode()

	<-ctx.Done()
	bdb.Close()
	err = s.Stop()
	if err != nil {
		log.Fatal(err)
	}
	os.RemoveAll(testdata.Node1Folder)
}

func runSecondNode() {
	os.Mkdir(testdata.Node2Folder, 0777)
	logfile, err := os.Create(testdata.Node2Folder + "/log.log")
	if err != nil {
		panic(err)
	}
	log := logger.New(logfile)
	log.WithDebug()
	ctx, cancel := context.WithCancel(context.Background())
	config.InterruptListener(log, cancel)

	bdb, err := bdb.NewBlockDB(testdata.Node2Folder, testdata.IntTestParams, log)
	if err != nil {
		log.Fatal(err)
	}

	c := testdata.Conf
	c.DataFolder = testdata.Node2Folder
	c.AddNodes = []peer.AddrInfo{hostMultiAddr}
	c.RPCPort = "24000"
	testServer, err := server.NewServer(ctx, &c, log, testdata.IntTestParams, bdb, initializationParams)
	if err != nil {
		log.Fatal(err)
	}
	go testServer.Start()

	<-ctx.Done()
	bdb.Close()
	err = testServer.Stop()
	if err != nil {
		log.Fatal(err)
	}
	os.RemoveAll(testdata.Node2Folder)
}