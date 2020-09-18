package chain_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/state"
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

var NumValidators = 10

type pair struct {
	public  *bls.PublicKey
	private *bls.SecretKey
}

var secrets = make([]pair, NumValidators)

var params = testdata.TestParams
var premineAddr = bls.RandKey()

var initParams state.InitializationParameters

func init() {
	_ = os.Remove("./chain.db")

	initParams = state.InitializationParameters{
		InitialValidators: make([]state.ValidatorInitialization, NumValidators),
		PremineAddress:    premineAddr.PublicKey().ToAccount(),
		GenesisTime:       time.Unix(time.Now().Unix()+5, 0),
	}

	for i := range secrets {
		key := bls.RandKey()
		secrets[i] = pair{
			public:  key.PublicKey(),
			private: key,
		}
		var pub [48]byte
		copy(pub[:], key.PublicKey().Marshal())
		initParams.InitialValidators[i] = state.ValidatorInitialization{
			PubKey:       hex.EncodeToString(key.PublicKey().Marshal()),
			PayeeAddress: premineAddr.PublicKey().ToAccount(),
		}
	}
}

func TestState(t *testing.T) {

	log := logger.New(os.Stdin)
	db, err := blockdb.NewBlockDB("./", params, log)
	assert.NoError(t, err)

	s, err := chain.NewStateService(log, initParams, params, db)
	assert.NoError(t, err)

	err = processBlock(s)
	assert.NoError(t, err)

	//i := 0
	//for {
	//	time.Sleep(time.Second * time.Duration(params.SlotDuration))
	//	err = processBlock(s)
	//	assert.NoError(t, err)
	//	if i == 50 {
	//		break
	//	}
	//	i++
	//}
}

func processBlock(ss chain.StateService) error {
	slot := getCurrentSlot() + 1
	if getCurrentSlot() == 0 {
		slot = 0
	}
	block, err := genNextBlock(ss, slot)
	if err != nil {
		return err
	}
	_, receipts, err := ss.Add(block)
	if err != nil {
		return err
	}
	if len(receipts) > 0 {
		msg := "\nEpoch Receipts\n----------\n"
		receiptTypes := make(map[string]int64)

		for _, r := range receipts {
			if _, ok := receiptTypes[r.TypeString()]; !ok {
				receiptTypes[r.TypeString()] = r.Amount
			} else {
				receiptTypes[r.TypeString()] += r.Amount
			}
		}

		for rt, amount := range receiptTypes {
			if amount > 0 {
				msg += fmt.Sprintf("rewarded %d for %s\n", amount, rt)
			} else if amount < 0 {
				msg += fmt.Sprintf("penalized %d for %s\n", -amount, rt)
			} else {
				msg += fmt.Sprintf("neutral increments for %s\n", rt)
			}
		}

		fmt.Println(msg)
	}
	return nil
}

func genNextBlock(ss chain.StateService, slot uint64) (*primitives.Block, error) {
	s, err := ss.TipStateAtSlot(slot)
	if err != nil {
		return nil, err
	}
	slotIndex := (slot + params.EpochLength - 1) % params.EpochLength
	proposerIndex := s.GetProposerQueue()[slotIndex]
	proposer := s.GetValidatorRegistry()[proposerIndex]

	var proposerKey *bls.SecretKey
	for _, p := range secrets {
		if bytes.Equal(proposer.PubKey[:], p.public.Marshal()) {
			proposerKey = p.private
		}
	}

	votes, err := getNextSlotVotes(ss, slot)

	block := &primitives.Block{
		Header: &primitives.BlockHeader{
			Version:       0,
			Nonce:         0,
			PrevBlockHash: ss.Tip().Hash,
			Timestamp:     uint64(time.Now().Unix()),
			Slot:          slot,
		},
		Votes: votes,
	}
	block.Header.VoteMerkleRoot = block.VotesMerkleRoot()
	block.Header.TxMerkleRoot = block.TransactionMerkleRoot()
	block.Header.TxMultiMerkleRoot = block.TransactionMultiMerkleRoot()
	block.Header.DepositMerkleRoot = block.DepositMerkleRoot()
	block.Header.ExitMerkleRoot = block.ExitMerkleRoot()
	block.Header.ProposerSlashingMerkleRoot = block.ProposerSlashingsRoot()
	block.Header.RANDAOSlashingMerkleRoot = block.RANDAOSlashingsRoot()
	block.Header.VoteSlashingMerkleRoot = block.VoteSlashingRoot()
	block.Header.GovernanceVotesMerkleRoot = block.GovernanceVoteMerkleRoot()

	blockHash := block.Hash()
	randaoHash := chainhash.HashH([]byte(fmt.Sprintf("%d", slot)))

	blockSig := proposerKey.Sign(blockHash[:])
	randaoSig := proposerKey.Sign(randaoHash[:])
	var sig, rs [96]byte
	copy(sig[:], blockSig.Marshal())
	copy(rs[:], randaoSig.Marshal())
	block.Signature = sig
	block.RandaoSignature = rs

	return block, nil
}

func getNextSlotVotes(ss chain.StateService, slot uint64) ([]*primitives.MultiValidatorVote, error) {

	s, err := ss.TipStateAtSlot(slot)
	if err != nil {
		return nil, err
	}

	toEpoch := (slot - 1) / params.EpochLength

	data := &primitives.VoteData{
		Slot:            slot,
		FromEpoch:       s.GetJustifiedEpoch(),
		FromHash:        s.GetJustifiedEpochHash(),
		ToEpoch:         toEpoch,
		ToHash:          s.GetRecentBlockHash(toEpoch*params.EpochLength-1, &params),
		BeaconBlockHash: ss.Tip().Hash,
		Nonce:           0,
	}

	dataHash := data.Hash()

	validators, err := s.GetVoteCommittee(slot, &params)
	if err != nil {
		return nil, err
	}

	var signatures []*bls.Signature

	bitlistVotes := bitfield.NewBitlist(uint64(len(validators)))

	validatorRegistry := s.GetValidatorRegistry()
	for i, index := range validators {
		validator := validatorRegistry[index]
		key := getValidatorKey(validator.PubKey[:])
		signatures = append(signatures, key.Sign(dataHash[:]))
		bitlistVotes.Set(uint(i))
	}
	var votes []*primitives.MultiValidatorVote
	if len(signatures) > 0 {
		sig := bls.AggregateSignatures(signatures)

		var voteSig [96]byte
		copy(voteSig[:], sig.Marshal())

		vote := &primitives.MultiValidatorVote{
			Data:                  data,
			ParticipationBitfield: bitlistVotes,
			Sig:                   voteSig,
		}
		votes = append(votes, vote)
	}
	return votes, nil
}

func getValidatorKey(pub []byte) *bls.SecretKey {
	for _, s := range secrets {
		if bytes.Equal(s.public.Marshal(), pub) {
			return s.private
		}
	}
	return nil
}

func getCurrentSlot() uint64 {
	slot := time.Now().Sub(initParams.GenesisTime) / (time.Duration(params.SlotDuration) * time.Second)
	if slot < 0 {
		return 0
	}
	return uint64(slot)
}
