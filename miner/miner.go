package miner

import (
	"bytes"
	"context"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/mempool"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/logger"
)

// Config is a config for the miner.
type Config struct {
	Log *logger.Logger
}

// BasicKeystore is a basic key store.
type BasicKeystore struct {
	keys map[[48]byte]bls.SecretKey
}

// GetKey gets the key for a certain validator.
func (b *BasicKeystore) GetKey(w *primitives.Validator) (*bls.SecretKey, bool) {
	var pub [48]byte
	copy(pub[:], w.PubKey)
	key, found := b.keys[pub]
	return &key, found
}

// NewBasicKeystore creates a key store from the following keys.
func NewBasicKeystore(keys []bls.SecretKey) *BasicKeystore {
	m := make(map[[48]byte]bls.SecretKey)
	for _, k := range keys {
		pubkey := k.PublicKey().Marshal()
		var pub [48]byte
		copy(pub[:], pubkey)
		m[pub] = k
	}

	return &BasicKeystore{
		keys: m,
	}
}

// Miner manages mining for the blockchain.
type Miner struct {
	log        *logger.Logger
	config     Config
	params     params.ChainParams
	chain      *chain.Blockchain
	keystore   *keystore.Keystore
	mineActive bool
	context    context.Context
	Stop       context.CancelFunc

	voteMempool    *mempool.VoteMempool
	coinsMempool   *mempool.CoinsMempool
	actionsMempool *mempool.ActionMempool

	blockTopic *pubsub.Topic
	voteTopic  *pubsub.Topic
}

// NewMiner creates a new miner from the parameters.
func NewMiner(config Config, params params.ChainParams, chain *chain.Blockchain, keystore *keystore.Keystore, hostnode *peers.HostNode, voteMempool *mempool.VoteMempool, coinsMempool *mempool.CoinsMempool, actionsMempool *mempool.ActionMempool) (miner *Miner, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	blockTopic, err := hostnode.Topic("blocks")
	if err != nil {
		cancel()
		return nil, err
	}
	voteTopic, err := hostnode.Topic("votes")
	if err != nil {
		cancel()
		return nil, err
	}
	miner = &Miner{
		log:            config.Log,
		config:         config,
		params:         params,
		chain:          chain,
		mineActive:     true,
		keystore:       keystore,
		context:        ctx,
		Stop:           cancel,
		voteMempool:    voteMempool,
		coinsMempool:   coinsMempool,
		actionsMempool: actionsMempool,

		blockTopic: blockTopic,
		voteTopic:  voteTopic,
	}
	chain.Notify(miner)
	return miner, nil
}

// NewTip implements the BlockchainNotifee interface.
func (m *Miner) NewTip(_ *index.BlockRow, block *primitives.Block, newState *primitives.State, _ []*primitives.EpochReceipt) {
	m.voteMempool.Remove(block)
	m.coinsMempool.RemoveByBlock(block)
	m.actionsMempool.RemoveByBlock(block, newState)
}

func (m *Miner) getCurrentSlot() uint64 {
	slot := int64(time.Now().Sub(m.chain.GenesisTime()) / (time.Duration(m.params.SlotDuration) * time.Second))
	if slot < 0 {
		return 0
	}
	return uint64(slot)
}

// getNextSlotTime gets the next slot time.
func (m *Miner) getNextBlockTime(nextSlot uint64) time.Time {
	return m.chain.GenesisTime().Add(time.Duration(nextSlot*m.params.SlotDuration) * time.Second)
}

// getNextSlotTime gets the next slot time.
func (m *Miner) getNextVoteTime(nextSlot uint64) time.Time {
	return m.chain.GenesisTime().Add(time.Duration(nextSlot*m.params.SlotDuration) * time.Second).Add(-time.Second * time.Duration(m.params.SlotDuration) / 2)
}

func (m *Miner) publishVote(vote *primitives.SingleValidatorVote) {
	buf := bytes.NewBuffer([]byte{})
	err := vote.Encode(buf)
	if err != nil {
		m.log.Errorf("error encoding vote: %s", err)
		return
	}

	if err := m.voteTopic.Publish(m.context, buf.Bytes()); err != nil {
		m.log.Errorf("error publishing vote: %s", err)
	}
}

func (m *Miner) publishBlock(block *primitives.Block) {
	buf := bytes.NewBuffer([]byte{})
	err := block.Encode(buf)
	if err != nil {
		m.log.Error(err)
		return
	}

	if err := m.blockTopic.Publish(m.context, buf.Bytes()); err != nil {
		m.log.Errorf("error publishing block: %s", err)
	}
}

// ProposerSlashingConditionViolated implements chain notifee.
func (m *Miner) ProposerSlashingConditionViolated(_ primitives.ProposerSlashing) {}

func (m *Miner) ProposeBlocks() {
	slotToPropose := m.getCurrentSlot() + 1

	blockTimer := time.NewTimer(time.Until(m.getNextBlockTime(slotToPropose)))

	for {
		select {
		case <-blockTimer.C:
			// check if we're an attester for this slot
			tip := m.chain.State().Tip()
			tipHash := tip.Hash

			state, err := m.chain.State().TipStateAtSlot(slotToPropose)
			if err != nil {
				m.log.Error(err)
				return
			}

			slotIndex := (slotToPropose + m.params.EpochLength - 1) % m.params.EpochLength

			proposerIndex := state.ProposerQueue[slotIndex]
			proposer := state.ValidatorRegistry[proposerIndex]

			if k, found := m.keystore.GetValidatorKey(proposer.PubKey); found {
				m.log.Infof("proposing for slot %d", slotToPropose)

				votes := m.voteMempool.Get(slotToPropose, state, &m.params, proposerIndex)

				depositTxs, state, err := m.actionsMempool.GetDeposits(int(m.params.MaxDepositsPerBlock), state)
				if err != nil {
					m.log.Error(err)
					return
				}

				coinTxs, state := m.coinsMempool.Get(m.params.MaxTxsPerBlock, state)

				exitTxs, err := m.actionsMempool.GetExits(int(m.params.MaxExitsPerBlock), state)
				if err != nil {
					m.log.Error(err)
					return
				}

				randaoSlashings, err := m.actionsMempool.GetRANDAOSlashings(int(m.params.MaxRANDAOSlashingsPerBlock), state)
				if err != nil {
					m.log.Error(err)
					return
				}

				voteSlashings, err := m.actionsMempool.GetVoteSlashings(int(m.params.MaxVoteSlashingsPerBlock), state)
				if err != nil {
					m.log.Error(err)
					return
				}

				proposerSlashings, err := m.actionsMempool.GetProposerSlashings(int(m.params.MaxProposerSlashingsPerBlock), state)
				if err != nil {
					m.log.Error(err)
					return
				}

				block := primitives.Block{
					Header: primitives.BlockHeader{
						Version:       0,
						Nonce:         0,
						PrevBlockHash: tipHash,
						Timestamp:     time.Now(),
						Slot:          slotToPropose,
					},
					Votes:             votes,
					Txs:               coinTxs,
					Deposits:          depositTxs,
					Exits:             exitTxs,
					RANDAOSlashings:   randaoSlashings,
					VoteSlashings:     voteSlashings,
					ProposerSlashings: proposerSlashings,
				}

				block.Header.VoteMerkleRoot = block.VotesMerkleRoot()
				block.Header.TxMerkleRoot = block.TransactionMerkleRoot()
				block.Header.DepositMerkleRoot = block.DepositMerkleRoot()
				block.Header.ExitMerkleRoot = block.ExitMerkleRoot()
				block.Header.ProposerSlashingMerkleRoot = block.ProposerSlashingsRoot()
				block.Header.RANDAOSlashingMerkleRoot = block.RANDAOSlashingsRoot()
				block.Header.VoteSlashingMerkleRoot = block.VoteSlashingRoot()
				block.Header.GovernanceVotesMerkleRoot = block.GovernanceVoteMerkleRoot()

				blockHash := block.Hash()
				randaoHash := chainhash.HashH([]byte(fmt.Sprintf("%d", slotToPropose)))

				blockSig := k.Sign(blockHash[:])
				randaoSig := k.Sign(randaoHash[:])

				block.Signature = blockSig.Marshal()
				block.RandaoSignature = randaoSig.Marshal()
				if err := m.chain.ProcessBlock(&block); err != nil {
					m.log.Error(err)
					return
				}

				go m.publishBlock(&block)
			}

			slotToPropose++
			blockTimer = time.NewTimer(time.Until(m.getNextBlockTime(slotToPropose)))
		case <-m.context.Done():
			m.log.Info("stopping miner")
			return
		}
	}
}

func (m *Miner) VoteForBlocks() {
	slotToVote := m.getCurrentSlot() + 1
	if slotToVote <= 0 {
		slotToVote = 1
	}

	voteTimer := time.NewTimer(time.Until(m.getNextVoteTime(slotToVote)))

	for {
		select {
		case <-voteTimer.C:
			// check if we're an attester for this slot
			m.log.Infof("sending votes for slot %d", slotToVote)

			s := m.chain.State()

			state, err := s.TipStateAtSlot(slotToVote)
			if err != nil {
				panic(err)
			}

			validators, err := state.GetVoteCommittee(slotToVote, &m.params)
			if err != nil {
				m.log.Errorf("error getting vote committee: %e", err)
				continue
			}
			toEpoch := (slotToVote - 1) / m.params.EpochLength

			beaconBlock, found := s.Chain().GetNodeBySlot(slotToVote - 1)
			if !found {
				panic("could not find block")
			}

			data := primitives.VoteData{
				Slot:            slotToVote,
				FromEpoch:       state.JustifiedEpoch,
				FromHash:        state.JustifiedEpochHash,
				ToEpoch:         toEpoch,
				ToHash:          state.GetRecentBlockHash(toEpoch*m.params.EpochLength-1, &m.params),
				BeaconBlockHash: beaconBlock.Hash,
			}

			dataHash := data.Hash()

			for i, validatorIdx := range validators {
				validator := state.ValidatorRegistry[validatorIdx]

				if k, found := m.keystore.GetValidatorKey(validator.PubKey); found {
					sig := k.Sign(dataHash[:])

					vote := primitives.SingleValidatorVote{
						Data:      data,
						Signature: *sig,
						Offset:    uint32(i),
						OutOf:     uint32(len(validators)),
					}

					m.voteMempool.Add(&vote)

					go m.publishVote(&vote)

					// DO NOT UNCOMMENT: slashing test
					//if validatorIdx == 0 {
					//	data2 := primitives.VoteData{
					//		Slot:            slotToVote,
					//		FromEpoch:       state.JustifiedEpoch,
					//		FromHash:        state.JustifiedEpochHash,
					//		ToEpoch:         toEpoch,
					//		ToHash:          state.GetRecentBlockHash(toEpoch*m.params.EpochLength-1, &m.params),
					//		BeaconBlockHash: chainhash.HashH([]byte("lol")),
					//	}
					//
					//	data2Hash := data2.Hash()
					//
					//	sig2 := k.Sign(data2Hash[:])
					//
					//	vote2 := primitives.SingleValidatorVote{
					//		Data:      data2,
					//		Signature: *sig2,
					//		Offset:    uint32(i),
					//		OutOf:     uint32(len(validators)),
					//	}
					//
					//	m.voteMempool.Add(&vote2)
					//
					//	go m.publishVote(&vote2)
					//}
				}
			}

			slotToVote++
			voteTimer = time.NewTimer(time.Until(m.getNextVoteTime(slotToVote)))
		case <-m.context.Done():
			m.log.Info("stopping voter")
			return
		}
	}
}

// Start runs the miner.
func (m *Miner) Start() error {
	numOurs := 0
	numTotal := 0
	for _, w := range m.chain.State().TipState().ValidatorRegistry {
		if _, ok := m.keystore.GetValidatorKey(w.PubKey); ok {
			numOurs++
		}
		numTotal++
	}

	m.log.Infof("starting miner with %d/%d active validators", numOurs, numTotal)

	go m.VoteForBlocks()
	go m.ProposeBlocks()

	return nil
}
