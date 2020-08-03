// +build backup_test

package backup_test

import (
	"context"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/server"
	"github.com/olympus-protocol/ogen/test"
	"github.com/olympus-protocol/ogen/utils/logger"
)

var premineAddr, _ = testdata.PremineAddr.PublicKey().ToAddress(testdata.IntTestParams.AddrPrefix.Public)

var initParams = primitives.InitializationParameters{
	GenesisTime:       time.Unix(time.Now().Unix()+30, 0),
	PremineAddress:    premineAddr,
	InitialValidators: []primitives.ValidatorInitialization{},
}

var F *server.Server
var FAddr peer.AddrInfo
var B *server.Server
var M *server.Server

// Backup Validator Test
// 1. Create two nodes that are creating a chain (Primary, Backup, Mover).
// The primary node shares the same keystore than the backup, the mover works alone.
// 2. The primary node should stop proposing blocks and the backup node should take the place of the primary node.
func TestMain(m *testing.M) {
	createValidators()
	// The mover node should be the first to initialize (create they keystore and include the validators to initParams)
	thirdNode()
	// The primary should be the second to initialize (to create the shared keystore)
	firstNode()
	// The backup node should load the initialization params for both previous nodes and connect to them
	secondNode()
	os.Exit(m.Run())
}

func createValidators() {
	// Create datafolder Primary Node
	err := os.Mkdir(testdata.Node1Folder, 0777)
	if err != nil {
		panic(err)
	}
	// Create datafolder Mover Node
	err = os.Mkdir(testdata.Node3Folder, 0777)
	if err != nil {
		panic(err)
	}

	var w sync.WaitGroup
	w.Add(2)
	go func(w *sync.WaitGroup) {
		keystorePrimary, err := keystore.NewKeystore(testdata.Node1Folder, nil, testdata.KeystorePass)
		if err != nil {
			panic(err)
		}
		// Generate the validators data.
		valDataPrimary, err := keystorePrimary.GenerateNewValidatorKey(128, testdata.KeystorePass)
		if err != nil {
			panic(err)
		}
		// Convert the validators to initialization params.
		for _, vk := range valDataPrimary {
			val := primitives.ValidatorInitialization{
				PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
				PayeeAddress: premineAddr,
			}
			initParams.InitialValidators = append(initParams.InitialValidators, val)
		}
		w.Done()
		return
	}(&w)

	go func(w *sync.WaitGroup) {
		keystoreMover, err := keystore.NewKeystore(testdata.Node3Folder, nil, testdata.KeystorePass)
		if err != nil {
			panic(err)
		}
		// Generate the validators data.
		valDataMover, err := keystoreMover.GenerateNewValidatorKey(128, testdata.KeystorePass)
		if err != nil {
			panic(err)
		}
		// Convert the validators to initialization params.
		for _, vk := range valDataMover {
			val := primitives.ValidatorInitialization{
				PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
				PayeeAddress: premineAddr,
			}
			initParams.InitialValidators = append(initParams.InitialValidators, val)
		}
		w.Done()
		return
	}(&w)
	w.Wait()
}

func firstNode() {
	// Create logger
	logfile, err := os.Create(testdata.Node1Folder + "/log.log")
	if err != nil {
		panic(err)
	}
	log := logger.New(logfile)
	log.WithDebug()

	// Load the block database
	db, err := bdb.NewBlockDB(testdata.Node1Folder, testdata.IntTestParams, log)
	if err != nil {
		log.Fatal(err)
	}

	// Load the conf params
	c := testdata.Conf

	// Override the datafolder.
	c.DataFolder = testdata.Node1Folder
	c.Port = "24132"

	// Create the server instance
	F, err = server.NewServer(context.Background(), &c, log, testdata.IntTestParams, db, initParams)
	if err != nil {
		log.Fatal(err)
	}

	FAddr = peer.AddrInfo{
		ID:    F.HostNode.GetHost().ID(),
		Addrs: F.HostNode.GetHost().Network().ListenAddresses(),
	}

	// Start the server
	go F.Start()

	// Open the keystore to start generating blocks
	err = F.Proposer.OpenKeystore(testdata.KeystorePass)
	if err != nil {
		log.Fatal(err)
	}

	// Start the proposer
	err = F.Proposer.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func secondNode() {
	// Create datafolder
	err := os.Mkdir(testdata.Node2Folder, 0777)
	if err != nil {
		panic(err)
	}
	// Create logger
	logfile, err := os.Create(testdata.Node2Folder + "/log.log")
	if err != nil {
		panic(err)
	}
	log := logger.New(logfile)
	log.WithDebug()

	// Copy the Node1 Keystore
	ks, err := ioutil.ReadFile(path.Join(testdata.Node1Folder, "keystore.db"))
	if err != nil {
		panic(err)
	}

	// Write the keystore db
	err = ioutil.WriteFile(path.Join(testdata.Node2Folder, "keystore.db"), ks, 0777)
	if err != nil {
		log.Fatal(err)
	}

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

	// Open the keystore to start generating blocks
	err = B.Proposer.OpenKeystore(testdata.KeystorePass)
	if err != nil {
		log.Fatal(err)
	}
}

func thirdNode() {
	// Create logger
	logfile, err := os.Create(testdata.Node3Folder + "/log.log")
	if err != nil {
		panic(err)
	}
	log := logger.New(logfile)
	log.WithDebug()

	// Load the block database
	db, err := bdb.NewBlockDB(testdata.Node3Folder, testdata.IntTestParams, log)
	if err != nil {
		log.Fatal(err)
	}

	// Load the conf params
	c := testdata.Conf

	// Override the datafolder.
	c.DataFolder = testdata.Node3Folder
	c.RPCPort = "25002"
	c.Port = "24130"
	// Create the server instance
	M, err = server.NewServer(context.Background(), &c, log, testdata.IntTestParams, db, initParams)
	if err != nil {
		log.Fatal(err)
	}

	// Start the server
	go M.Start()

	// Open the keystore to start generating blocks
	err = M.Proposer.OpenKeystore(testdata.KeystorePass)
	if err != nil {
		log.Fatal(err)
	}

	// Start the proposer
	err = M.Proposer.Start()
	if err != nil {
		log.Fatal(err)
	}
}

type blockNotifee struct {
	slash chan struct{}
}

func newBlockNotifee(ctx context.Context, chain *chain.Blockchain) blockNotifee {
	bn := blockNotifee{
		slash: make(chan struct{}),
	}
	go func() {
		chain.Notify(&bn)
		<-ctx.Done()
		chain.Unnotify(&bn)
	}()

	return bn
}

func (bn *blockNotifee) NewTip(row *index.BlockRow, block *primitives.Block, newState *primitives.State, receipts []*primitives.EpochReceipt) {}

func (bn *blockNotifee) ProposerSlashingConditionViolated(slashing *primitives.ProposerSlashing) {
	bn.slash <-struct {}{}
}

// Since nodes are not connect all 3 tip states should be the same.
func Test_StallProposing(t *testing.T) {
	tipPrimary := F.Chain.State().Tip()
	tipMover := M.Chain.State().Tip()
	tipBackup := M.Chain.State().Tip()
	assert.Equal(t, tipPrimary, tipMover, tipBackup)
}

func Test_Connections(t *testing.T) {
	// The backup node should connect to the Primary node
	err := B.HostNode.GetHost().Connect(context.TODO(), FAddr)
	assert.NoError(t, err)
	// The mover node should connect to the Primary node
	err = M.HostNode.GetHost().Connect(context.TODO(), FAddr)
	assert.NoError(t, err)
}

// Start the backup proposing routine with a delay from the first node. If starting at the same time it will get slashed.
func Test_StartBackupProposing(t *testing.T) {
	time.Sleep(time.Second * 15)
	err := B.Proposer.Start()
	assert.NoError(t, err)
	time.Sleep(time.Second * 15)
	go func() {
		bn := newBlockNotifee(context.Background(), M.Chain)
		for {
			<-bn.slash
			assert.Fail(t, "slashing detected, backup failed")
		}
	}()
}

// Stop the primary node proposing routine and check the backup is doing the voting/proposing job.
func Test_StopPrimaryProposer(t *testing.T) {
	F.Proposer.Stop()
	time.Sleep(time.Minute)
}
