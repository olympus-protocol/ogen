// +build mempools_test

package mempools_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/server"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
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
	GenesisTime:       time.Unix(time.Now().Unix(), 0),
	PremineAddress:    premineAddr,
	InitialValidators: []primitives.ValidatorInitialization{},
}

var valDataPrimary []bls_interface.SecretKey
var newValidator bls_interface.SecretKey

var F *server.Server
var B *server.Server

// Mempools Test
// 1. Create a single node instance
// 2. Fill

func TestMain(m *testing.M) {
	// remove data folders if it exists
	_ = os.RemoveAll(testdata.Node1Folder)
	_ = os.RemoveAll(testdata.Node2Folder)
	// Create the validators
	createValidators()
	// Run first node.
	firstNode()
	// Start secondary node
	//secondNode()
	// Run tests
	os.Exit(m.Run())
}

func createValidators() {
	// Create datafolder Primary Node
	_ = os.Mkdir(testdata.Node1Folder, 0777)

	// Create the keystore
	k := keystore.NewKeystore(testdata.Node1Folder, nil)

	// Generate the validators data.
	valDataPrimary, err := k.GenerateNewValidatorKey(32)
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
	err = k.Close()
	if err != nil {
		panic(err)
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
	db, err := blockdb.NewBlockDB(testdata.Node1Folder, testdata.IntTestParams, log)
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

	// Start the server
	go F.Start()

	err = F.Proposer.OpenKeystore()
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
	db, err := blockdb.NewBlockDB(testdata.Node2Folder, testdata.IntTestParams, log)
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

var votes []*primitives.MultiValidatorVote

// Create the votes and add it to the mempool with validation.
func TestVoteMempool_AddValidate(t *testing.T) {
	slotToVote := F.Chain.State().TipState().Slot + 1

	state, err := F.Chain.State().TipStateAtSlot(slotToVote)
	assert.NoError(t, err)

	// Get validator indices tha should vote for next slot.
	votesIdx, err := state.GetVoteCommittee(slotToVote, &testdata.IntTestParams)
	assert.NoError(t, err)

	toEpoch := (slotToVote - 1) / testdata.IntTestParams.EpochLength
	beaconBlock, found := F.Chain.State().Chain().GetNodeBySlot(slotToVote - 1)
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

		sec, ok := F.Proposer.Keystore().GetValidatorKey(val.PubKey)
		assert.True(t, ok)
		assert.NotNil(t, sec)

		sig := sec.Sign(dataHash[:])

		var sigB [96]byte
		copy(sigB[:], sig.Marshal())
		v := &primitives.MultiValidatorVote{
			Data: &voteData,
			Sig:  sigB,
			//Offset: uint64(i),
			//OutOf:  uint64(len(votesIdx)),
		}

		votes = append(votes, v)
		err := F.Mempools.Votes.AddValidate(v, state)
		assert.NoError(t, err)
	}

}

var aggVote = new(primitives.MultiValidatorVote)

func TestVoteAggregation(t *testing.T) {
	slotToVote := F.Chain.State().TipState().Slot + 1

	state, err := F.Chain.State().TipStateAtSlot(slotToVote)
	assert.NoError(t, err)

	// This assumes all votes data is the same
	aggVote.Data = votes[0].Data
	sigs := make([]bls_interface.Signature, 0)

	for _, v := range votes {
		// Test vote validity without aggregation
		assert.NoError(t, state.IsVoteValid(v, &testdata.IntTestParams))

		sig, err := v.Signature()

		assert.NoError(t, err)

		sigs = append(sigs, sig)
	}

	sig := bls.AggregateSignatures(sigs)
	var sigB [96]byte
	copy(sigB[:], sig.Marshal())
	aggVote.Sig = sigB

	// Create a list bitfield list with the amount of validators voting
	aggVote.ParticipationBitfield = bitfield.NewBitlist(uint64(len(votes)) + 7)

	// Mark each bitfield with the validator index
	for _, v := range votes {
		aggVote.ParticipationBitfield.Set(uint(v.Offset))
	}
	assert.NoError(t, state.IsVoteValid(aggVote, &testdata.IntTestParams))
}

func TestVoteMempool_Get(t *testing.T) {

	slotToPropose := F.Chain.State().TipState().Slot + 2

	state, err := F.Chain.State().TipStateAtSlot(slotToPropose)
	assert.NoError(t, err)

	slotIndex := (slotToPropose + testdata.IntTestParams.EpochLength - 1) % testdata.IntTestParams.EpochLength

	proposerIndex := state.ProposerQueue[slotIndex]

	votes, err := F.Mempools.Votes.Get(slotToPropose, state, &testdata.IntTestParams, proposerIndex)

	newState, err := F.Chain.State().TipStateAtSlot(F.Chain.State().TipState().Slot + 1)
	assert.NoError(t, err)

	assert.NoError(t, newState.IsVoteValid(votes[0], &testdata.IntTestParams))

	assert.NoError(t, err)

	b1, err := aggVote.Marshal()
	assert.NoError(t, err)

	b2, err := votes[0].Marshal()
	assert.NoError(t, err)

	var mv1 = new(primitives.MultiValidatorVote)
	var mv2 = new(primitives.MultiValidatorVote)

	err = mv1.Unmarshal(b1)
	assert.NoError(t, err)

	err = mv2.Unmarshal(b2)
	assert.NoError(t, err)

	assert.Equal(t, aggVote, votes[0])

}

func TestActionMempool_Deposits(t *testing.T) {
	priv := testdata.PremineAddr
	pub := priv.PublicKey()

	validatorPrivs, err := F.Proposer.Keystore().GenerateNewValidatorKey(1)
	assert.NoError(t, err)

	validatorPriv := validatorPrivs[0]
	newValidator = validatorPriv

	validatorPub := validatorPriv.PublicKey()
	validatorPubBytes := validatorPub.Marshal()
	validatorPubHash := chainhash.HashH(validatorPubBytes[:])

	validatorProofOfPossession := validatorPriv.Sign(validatorPubHash[:])

	addr, err := pub.Hash()
	assert.NoError(t, err)
	var p [48]byte
	var s [96]byte
	copy(p[:], validatorPubBytes)
	copy(s[:], validatorProofOfPossession.Marshal())
	depositData := &primitives.DepositData{
		PublicKey:         p,
		ProofOfPossession: s,
		WithdrawalAddress: addr,
	}

	buf, err := depositData.Marshal()
	assert.NoError(t, err)

	depositHash := chainhash.HashH(buf)

	depositSig := priv.Sign(depositHash[:])

	var pubKey [48]byte
	var ds [96]byte
	copy(pubKey[:], pub.Marshal())
	copy(ds[:], depositSig.Marshal())
	deposit := &primitives.Deposit{
		PublicKey: pubKey,
		Signature: ds,
		Data:      depositData,
	}

	state := F.Chain.State().TipState()

	err = F.Mempools.Actions.AddDeposit(deposit, state)
	assert.NoError(t, err)

	// there should be one deposit in the mempool
	depositTxs, _, err := F.Mempools.Actions.GetDeposits(1, state)
	assert.NoError(t, err)

	// assert that it is the same deposit
	assert.Equal(t, 1, len(depositTxs))
	assert.Equal(t, deposit.PublicKey, depositTxs[0].PublicKey)
	assert.Equal(t, deposit.Signature, depositTxs[0].Signature)
	/*err = F.Proposer.Start()
	assert.NoError(t, err)
	// give some to the block
	time.Sleep(time.Second * 20)*/

}

func TestActionMempool_ExitDeposits(t *testing.T) {
	priv := testdata.PremineAddr
	pub := priv.PublicKey()

	validatorPub := valDataPrimary[0].PublicKey()
	msg := fmt.Sprintf("exit %x", validatorPub.Marshal())
	msgHash := chainhash.HashH([]byte(msg))

	sig := priv.Sign(msgHash[:])
	var valp, withp [48]byte
	var s [96]byte
	copy(valp[:], validatorPub.Marshal())
	copy(withp[:], pub.Marshal())
	copy(s[:], sig.Marshal())
	exit := &primitives.Exit{
		ValidatorPubkey: valp,
		WithdrawPubkey:  withp,
		Signature:       s,
	}

	state := F.Chain.State().TipState()
	err := F.Mempools.Actions.AddExit(exit, state)
	assert.NoError(t, err)

	// there should be one exitdeposit in the mempool
	exitTxs, err := F.Mempools.Actions.GetExits(1, state)
	assert.NoError(t, err)

	// assert that it is the same deposit
	assert.Equal(t, 1, len(exitTxs))
	assert.Equal(t, exit.WithdrawPubkey, exitTxs[0].WithdrawPubkey)
	assert.Equal(t, exit.Signature, exitTxs[0].Signature)

}
