// +build chain_test

package chain_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/server"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/olympus-protocol/ogen/utils/logger"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
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

var runUntilHeight = 20

// Chain Test
// 1. This test will create a node that moves the chain based on the initialization params.
// 2. The second node should follow the chain.
// 3. The chain will finish once reached the 1000 slot.
func TestMain(m *testing.M) {
	// Create the validators
	createValidators()
	// Run first node.
	firstNode()
	// Run second node.
	secondNode()
	// Log epoch receipts from first node
	go logNotify()
	// Run tests
	os.Exit(m.Run())
}

func createValidators() {
	// Create datafolder Primary Node
	_ = os.Mkdir(testdata.Node1Folder, 0777)

	// Create the keystore
	k, err := keystore.NewKeystore(testdata.Node1Folder, nil, testdata.KeystorePass)
	if err != nil {
		panic(err)
	}

	// Generate the validators data.
	valDataPrimary, err := k.GenerateNewValidatorKey(8, testdata.KeystorePass)
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

type blockNotifee struct {
	blocks chan blockAndReceipts
}

type blockAndReceipts struct {
	block    *primitives.Block
	receipts []*primitives.EpochReceipt
	state    *primitives.State
}

func newBlockNotifee(ctx context.Context, chain *chain.Blockchain) blockNotifee {
	bn := blockNotifee{
		blocks: make(chan blockAndReceipts),
	}
	go func() {
		chain.Notify(&bn)
		<-ctx.Done()
		chain.Unnotify(&bn)
	}()

	return bn
}

func (bn *blockNotifee) NewTip(row *index.BlockRow, block *primitives.Block, newState *primitives.State, receipts []*primitives.EpochReceipt) {
	fmt.Printf("Slot %v Hash: %s Height: %v StateRoot: %s \n", row.Slot, hex.EncodeToString(row.Hash[:]), row.Height, hex.EncodeToString(row.StateRoot[:]))
	fmt.Printf("%v Epoch Receipts \n", len(receipts))
	for _, receipt := range receipts {
		fmt.Printf("Validator: %v Amount: %v Type: %s \n", receipt.Validator, receipt.Amount, receipt.TypeString())
	}
}

func (bn *blockNotifee) ProposerSlashingConditionViolated(slashing *primitives.ProposerSlashing) {
	fmt.Printf("Slashing:  %s \n", hex.EncodeToString(slashing.ValidatorPublicKey[:]))
}


func logNotify() {
	bn := newBlockNotifee(context.Background(), F.Chain)
	for {
		select {
		case bl := <-bn.blocks:
			fmt.Println(bl)
		}
	}

}

func Test_Connections(t *testing.T) {
	// The backup node should connect to the first node
	err := B.HostNode.GetHost().Connect(context.TODO(), FAddr)
	assert.NoError(t, err)
}

func Test_ReachHeight(t *testing.T) {
	// Check until the node reaches runUntilHeight
	// Include a stall detector.
	stall := 0
	height := F.Chain.State().Height()
	for {
		if F.Chain.State().Tip().Height >= uint64(runUntilHeight) {
			break
		}
		time.Sleep(time.Second)
		if F.Chain.State().Height() == height {
			stall++
		} else {
			height = F.Chain.State().Height()
		}
		if stall >= 30 {
			assert.Fail(t, "proposer stalled")
		}
	}
}

func Test_NodesStateMatch(t *testing.T) {
	// Stop proposing new blocks
	F.Proposer.Stop()

	// State from both nodes should match.
	assert.Equal(t, F.Chain.State().Tip().Height, B.Chain.State().Tip().Height)
	assert.Equal(t, F.Chain.State().Tip().Hash, B.Chain.State().Tip().Hash)
}

func Test_JustifiedHash(t *testing.T) {
	// Since it is only one node, the latest justified state should be slot
}

func Test_FinalizedHash(t *testing.T) {
	// Since it is only one node, the latest finalized state should be slot
}