package mempools_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/server"
	testdata "github.com/olympus-protocol/ogen/test"
	bitfcheck "github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/olympus-protocol/ogen/utils/logger"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func init() {
	err := bls.Initialize(testdata.IntTestParams)
	if err != nil {
		panic(err)
	}
}

var premineAddr, _ = testdata.PremineAddr.PublicKey().ToAccount()

var initParams = primitives.InitializationParameters{
	GenesisTime:       time.Unix(time.Now().Unix()+10, 0),
	PremineAddress:    premineAddr,
	InitialValidators: []primitives.ValidatorInitialization{},
}

var F *server.Server
var FAddr peer.AddrInfo
var B *server.Server

// Mempools Test
// 1. Create two nodes and connect to produce some blocks.
// 2. Stop the block producer and run the test framework over the mempools.

func TestMain(m *testing.M) {
	// Create the validators
	createValidators()
	// Run first node.
	firstNode()
	// Run second node.
	secondNode()
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
	valDataPrimary, err := k.GenerateNewValidatorKey(32, testdata.KeystorePass)
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

func Test_Connections(t *testing.T) {
	// The backup node should connect to the first node
	err := B.HostNode.GetHost().Connect(context.TODO(), FAddr)
	assert.NoError(t, err)
}

func Test_StopProposer(t *testing.T) {
	time.Sleep(time.Second * 40)
	F.Proposer.Stop()
}

func Test_CleanMempools(t *testing.T) {
	F.Mempools.Votes.Clear()
	F.Mempools.Actions.Clear()
	F.Mempools.Coins.Clear()
}

var votes []*primitives.SingleValidatorVote

func getSlot() uint64 {
	slot := time.Now().Sub(F.Chain.GenesisTime()) / (time.Duration(testdata.IntTestParams.SlotDuration) * time.Second)
	if slot < 0 {
		return 0
	}
	return uint64(slot)
}

// Create the votes and add it to the mempool with validation.
func TestVoteMempool_AddValidate(t *testing.T) {
	slotToVote := getSlot() + 1

	state, err := F.Chain.State().TipStateAtSlot(slotToVote)
	assert.NoError(t, err)

	// Get validator indices tha should vote for next slot.
	votesIdx, err := state.GetVoteCommittee(slotToVote, &testdata.IntTestParams)
	assert.NoError(t, err)

	toEpoch := (slotToVote - 1) / testdata.IntTestParams.EpochLength
	beaconBlock, found := F.Chain.State().Chain().GetNodeBySlot(slotToVote)
	if !found {
		panic("could not find block")
	}

	voteData := primitives.VoteData{
		Slot:            slotToVote,
		FromEpoch:       state.JustifiedEpoch,
		FromHash:        state.JustifiedEpochHash,
		ToEpoch:         toEpoch,
		ToHash:          state.GetRecentBlockHash(toEpoch*testdata.IntTestParams.EpochLength-1, &testdata.IntTestParams),
		BeaconBlockHash: beaconBlock.Hash,
		Nonce:           0,
	}

	dataHash := voteData.Hash()

	for i, idx := range votesIdx {

		val := state.ValidatorRegistry[idx]

		sec, ok := F.Proposer.Keystore.GetValidatorKey(val.PubKey)
		assert.True(t, ok)
		assert.NotNil(t, sec)

		sig := sec.Sign(dataHash[:])

		var sigB [96]byte
		copy(sigB[:], sig.Marshal())
		v := &primitives.SingleValidatorVote{
			Data:   &voteData,
			Sig:    sigB,
			Offset: uint64(i),
			OutOf:  uint64(len(votesIdx)),
		}
		votes = append(votes, v)
		err := F.Mempools.Votes.AddValidate(v, state)
		assert.NoError(t, err)
	}

}

func TestVoteAggregation(t *testing.T) {
	slotToVote := getSlot() + 1

	state, err := F.Chain.State().TipStateAtSlot(slotToVote)
	assert.NoError(t, err)

	mv := new(primitives.MultiValidatorVote)

	// This assumes all votes data is the same
	mv.Data = votes[0].Data
	sigs := make([]*bls.Signature, 0)

	for _, v := range votes {
		// Test vote validity without aggregation
		multi := v.AsMulti()
		assert.NoError(t, state.IsVoteValid(multi, &testdata.IntTestParams))

		sig, err := v.Signature()

		assert.NoError(t, err)

		sigs = append(sigs, sig)
	}

	sig := bls.AggregateSignatures(sigs)
	var sigB [96]byte
	copy(sigB[:], sig.Marshal())
	mv.Sig = sigB

	// Create a list bitfield list with the amount of validators voting
	mv.ParticipationBitfield = bitfield.NewBitlist(uint64(len(votes)))

	// Mark each bitfield with the validator index
	for _, v := range votes {
		bitfcheck.Set(mv.ParticipationBitfield, uint(v.Offset))
	}
	assert.NoError(t, state.IsVoteValid(mv, &testdata.IntTestParams))
}

func TestVoteMempool_Get(t *testing.T) {

	slotToPropose := getSlot() + 1

	state, err := F.Chain.State().TipStateAtSlot(slotToPropose)

	assert.NoError(t, err)

	slotIndex := (slotToPropose + testdata.IntTestParams.EpochLength - 1) % testdata.IntTestParams.EpochLength

	proposerIndex := state.ProposerQueue[slotIndex]

	votes, err := F.Mempools.Votes.Get(slotToPropose, state, &testdata.IntTestParams, proposerIndex)

	assert.NoError(t, err)

	fmt.Println(votes)
}

// Vote Mempools Test
