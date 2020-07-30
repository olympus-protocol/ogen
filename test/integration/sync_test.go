// +build sync_test

package sync_test

import (
	"context"
	"encoding/hex"
	"fmt"
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
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/logger"
)

var s *server.Server
var ps *server.Server

// Sync test.
// 1. The initial node will load pre-built chain containing 985 blocks at a certain genesis time.
// 2. The second node will connect to the initial node and sync those 985 blocks.
// 3. A third process will check the second node with timeouts to see if the sync is stalled at a certain point of the sync.
// 	  if the second node finishes the sync of the 985 blocks the test pass.
// There is a workaround to fetch the initilization params since those were not store during the generation process.
func TestMain(m *testing.M) {
	// Create datafolder
	os.Mkdir(testdata.Node1Folder, 0777)

	// Create logger
	logfile, err := os.Create(testdata.Node1Folder + "/log.log")
	if err != nil {
		panic(err)
	}
	log := logger.New(logfile)
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

	validators := []primitives.ValidatorInitialization{}
	for _, vk := range validatorKeys {
		val := primitives.ValidatorInitialization{
			PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
			PayeeAddress: bech32.Encode(testdata.IntTestParams.AddrPrefix.Public, []byte{163, 15, 14, 86, 107, 205, 124, 126, 243, 101, 198, 2, 95, 29, 158, 221, 60, 108, 201, 78}),
		}
		validators = append(validators, val)
	}

	// Create the initialization parameters
	ip := primitives.InitializationParameters{
		GenesisTime:       time.Unix(1595126983, 0),
		PremineAddress:    bech32.Encode(testdata.IntTestParams.AddrPrefix.Public, []byte{163, 15, 14, 86, 107, 205, 124, 126, 243, 101, 198, 2, 95, 29, 158, 221, 60, 108, 201, 78}),
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
	ps, err = server.NewServer(ctx, &c, log, testdata.IntTestParams, bdb, ip)
	if err != nil {
		log.Fatal(err)
	}
	go ps.Start()
	os.Exit(m.Run())
	var initialValidators []primitives.ValidatorInitialization
	for _, sv := range s.Chain.State().TipState().ValidatorRegistry {
		initialValidators = append(initialValidators, primitives.ValidatorInitialization{
			PubKey:       hex.EncodeToString(sv.PubKey),
			PayeeAddress: bech32.Encode(testdata.IntTestParams.AddrPrefix.Public, sv.PayeeAddress[:]),
		})
	}
	ip.InitialValidators = initialValidators
	go runSecondNode(s, ip)
	<-ctx.Done()
	bdb.Close()
	err = ps.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func runSecondNode(ps *server.Server, ip primitives.InitializationParameters) {
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
	psAddr := peer.AddrInfo{
		ID:    ps.HostNode.GetHost().ID(),
		Addrs: ps.HostNode.GetHost().Network().ListenAddresses(),
	}
	c := testdata.Conf
	c.DataFolder = testdata.Node2Folder
	c.AddNodes = []peer.AddrInfo{psAddr}
	c.RPCPort = "24000"
	s, err = server.NewServer(ctx, &c, log, testdata.IntTestParams, bdb, ip)
	if err != nil {
		log.Fatal(err)
	}
	go s.Start()
	<-ctx.Done()
	bdb.Close()
	err = s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func Test_SyncStatus(t *testing.T) {
	stall := 0
	height := s.Chain.State().Height()
	for {
		if ps.Chain.State().Tip().Hash.IsEqual(&s.Chain.State().Tip().Hash) {
			os.Exit(0)
		}
		time.Sleep(time.Second)
		if s.Chain.State().Height() == height {
			stall++
		} else {
			height = s.Chain.State().Height()
		}
		if stall >= 30 {
			fmt.Println("test failed - stall time exceed")
			os.Exit(0)
		}
	}
}
