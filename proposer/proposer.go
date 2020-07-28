package proposer

import (
	"context"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

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

// Config is a config for the proposer.
type Config struct {
	Datadir string
	Log     *logger.Logger
}

// Proposer manages mining for the blockchain.
type Proposer struct {
	log        *logger.Logger
	config     Config
	params     params.ChainParams
	chain      *chain.Blockchain
	Keystore   *keystore.Keystore
	mineActive bool
	context    context.Context
	Stop       context.CancelFunc

	voteMempool    *mempool.VoteMempool
	coinsMempool   *mempool.CoinsMempool
	actionsMempool *mempool.ActionMempool
	hostnode       *peers.HostNode
	blockTopic     *pubsub.Topic
	voteTopic      *pubsub.Topic
}

// OpenKeystore opens the keystore with the provided password
func (p *Proposer) OpenKeystore(password string) (err error) {
	p.Keystore, err = keystore.NewKeystore(p.config.Datadir, p.log, password)
	if err != nil {
		return err
	}
	return nil
}

// NewProposer creates a new proposer from the parameters.
func NewProposer(config Config, params params.ChainParams, chain *chain.Blockchain, hostnode *peers.HostNode, voteMempool *mempool.VoteMempool, coinsMempool *mempool.CoinsMempool, actionsMempool *mempool.ActionMempool) (proposer *Proposer, err error) {
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
	proposer = &Proposer{
		log:            config.Log,
		config:         config,
		params:         params,
		chain:          chain,
		mineActive:     true,
		context:        ctx,
		Stop:           cancel,
		voteMempool:    voteMempool,
		coinsMempool:   coinsMempool,
		actionsMempool: actionsMempool,
		hostnode:       hostnode,
		blockTopic:     blockTopic,
		voteTopic:      voteTopic,
	}
	chain.Notify(proposer)
	return proposer, nil
}

// NewTip implements the BlockchainNotifee interface.
func (p *Proposer) NewTip(_ *index.BlockRow, block *primitives.Block, newState *primitives.State, _ []*primitives.EpochReceipt) {
	p.voteMempool.Remove(block)
	p.coinsMempool.RemoveByBlock(block)
	p.actionsMempool.RemoveByBlock(block, newState)
}

func (p *Proposer) getCurrentSlot() uint64 {
	slot := time.Now().Sub(p.chain.GenesisTime()) / (time.Duration(p.params.SlotDuration) * time.Second)
	if slot < 0 {
		return 0
	}
	return uint64(slot)
}

// getNextSlotTime gets the next slot time.
func (p *Proposer) getNextBlockTime(nextSlot uint64) time.Time {
	return p.chain.GenesisTime().Add(time.Duration(nextSlot*p.params.SlotDuration) * time.Second)
}

// getNextSlotTime gets the next slot time.
func (p *Proposer) getNextVoteTime(nextSlot uint64) time.Time {
	return p.chain.GenesisTime().Add(time.Duration(nextSlot*p.params.SlotDuration) * time.Second).Add(-time.Second * time.Duration(p.params.SlotDuration) / 2)
}

func (p *Proposer) publishVote(vote *primitives.SingleValidatorVote) {
	buf, err := vote.Marshal()
	if err != nil {
		p.log.Errorf("error encoding vote: %s", err)
		return
	}

	if err := p.voteTopic.Publish(p.context, buf); err != nil {
		p.log.Errorf("error publishing vote: %s", err)
	}
}

func (p *Proposer) publishBlock(block *primitives.Block) {
	buf, err := block.Marshal()
	if err != nil {
		p.log.Error(err)
		return
	}

	if err := p.blockTopic.Publish(p.context, buf); err != nil {
		p.log.Errorf("error publishing block: %s", err)
	}
}

// ProposerSlashingConditionViolated implements chain notifee.
func (p *Proposer) ProposerSlashingConditionViolated(_ *primitives.ProposerSlashing) {}

func (p *Proposer) ProposeBlocks() {
	slotToPropose := p.getCurrentSlot() + 1

	blockTimer := time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))

	for {
		select {
		case <-blockTimer.C:
			if p.hostnode.PeersConnected() == 0 && p.hostnode.Syncing() {
				p.log.Infof("blockchain not synced... trying to mine in 10 seconds")
				blockTimer = time.NewTimer(time.Second * 10)
				continue
			}
			//if p.chain.State().Tip().Slot+p.params.EpochLength < slotToPropose {
			//	p.log.Infof("blockchain not synced... trying to mine in 10 seconds")

			// wait 10 seconds before starting the next vote
			//	blockTimer = time.NewTimer(time.Second * 10)
			//	continue
			//}

			// check if we're an attester for this slot
			tip := p.chain.State().Tip()
			tipHash := tip.Hash

			state, err := p.chain.State().TipStateAtSlot(slotToPropose)
			if err != nil {
				p.log.Error(err)
				return
			}

			slotIndex := (slotToPropose + p.params.EpochLength - 1) % p.params.EpochLength

			proposerIndex := state.ProposerQueue[slotIndex]
			proposer := state.ValidatorRegistry[proposerIndex]

			if k, found := p.Keystore.GetValidatorKey(proposer.PubKey); found {
				p.log.Infof("proposing for slot %d", slotToPropose)

				votes, err := p.voteMempool.Get(slotToPropose, state, &p.params, proposerIndex)
				if err != nil {
					p.log.Error(err)
					return
				}
				depositTxs, state, err := p.actionsMempool.GetDeposits(int(p.params.MaxDepositsPerBlock), state)
				if err != nil {
					p.log.Error(err)
					return
				}

				coinTxs, state := p.coinsMempool.Get(p.params.MaxTxsPerBlock, state)

				exitTxs, err := p.actionsMempool.GetExits(int(p.params.MaxExitsPerBlock), state)
				if err != nil {
					p.log.Error(err)
					return
				}

				randaoSlashings, err := p.actionsMempool.GetRANDAOSlashings(int(p.params.MaxRANDAOSlashingsPerBlock), state)
				if err != nil {
					p.log.Error(err)
					return
				}

				voteSlashings, err := p.actionsMempool.GetVoteSlashings(int(p.params.MaxVoteSlashingsPerBlock), state)
				if err != nil {
					p.log.Error(err)
					return
				}

				proposerSlashings, err := p.actionsMempool.GetProposerSlashings(int(p.params.MaxProposerSlashingsPerBlock), state)
				if err != nil {
					p.log.Error(err)
					return
				}

				block := primitives.Block{
					Header: &primitives.BlockHeader{
						Version:       0,
						Nonce:         0,
						PrevBlockHash: tipHash,
						Timestamp:     uint64(time.Now().Unix()),
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
				var s, rs [96]byte
				copy(s[:], blockSig.Marshal())
				copy(rs[:], randaoSig.Marshal())
				block.Signature = s
				block.RandaoSignature = rs
				if err := p.chain.ProcessBlock(&block); err != nil {
					p.log.Error(err)
					return
				}

				go p.publishBlock(&block)
			}

			slotToPropose++
			blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
		case <-p.context.Done():
			p.log.Info("stopping proposer")
			return
		}
	}
}

func (p *Proposer) VoteForBlocks() {
	slotToVote := p.getCurrentSlot() + 1
	if slotToVote <= 0 {
		slotToVote = 1
	}

	voteTimer := time.NewTimer(time.Until(p.getNextVoteTime(slotToVote)))

	for {
		select {
		case <-voteTimer.C:
			// check if we're an attester for this slot
			p.log.Infof("sending votes for slot %d", slotToVote)

			if p.hostnode.PeersConnected() == 0 && p.hostnode.Syncing() {
				voteTimer = time.NewTimer(time.Second * 10)
				p.log.Infof("blockchain not synced... trying to mine in 10 seconds")
				continue
			}
			//if p.chain.State().Tip().Slot+p.params.EpochLength < slotToVote {
			//	p.log.Infof("blockchain not synced... trying to mine in 10 seconds")

			// wait 10 seconds before starting the next vote
			//	voteTimer = time.NewTimer(time.Second * 10)
			//	continue
			//}

			s := p.chain.State()

			state, err := s.TipStateAtSlot(slotToVote)
			if err != nil {
				panic(err)
			}

			validators, err := state.GetVoteCommittee(slotToVote, &p.params)
			if err != nil {
				p.log.Errorf("error getting vote committee: %e", err)
				continue
			}
			toEpoch := (slotToVote - 1) / p.params.EpochLength

			beaconBlock, found := s.Chain().GetNodeBySlot(slotToVote - 1)
			if !found {
				panic("could not find block")
			}

			data := primitives.VoteData{
				Slot:            slotToVote,
				FromEpoch:       state.JustifiedEpoch,
				FromHash:        state.JustifiedEpochHash,
				ToEpoch:         toEpoch,
				ToHash:          state.GetRecentBlockHash(toEpoch*p.params.EpochLength-1, &p.params),
				BeaconBlockHash: beaconBlock.Hash,
			}

			dataHash := data.Hash()

			for i, validatorIdx := range validators {
				validator := state.ValidatorRegistry[validatorIdx]

				if k, found := p.Keystore.GetValidatorKey(validator.PubKey); found {
					sig := k.Sign(dataHash[:])
					var s [96]byte
					copy(s[:], sig.Marshal())
					vote := primitives.SingleValidatorVote{
						Data:   &data,
						Sig:    s,
						Offset: uint64(i),
						OutOf:  uint64(len(validators)),
					}

					p.voteMempool.Add(&vote)

					go p.publishVote(&vote)

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
			voteTimer = time.NewTimer(time.Until(p.getNextVoteTime(slotToVote)))
		case <-p.context.Done():
			p.log.Info("stopping voter")
			return
		}
	}
}

// Start runs the proposer.
func (p *Proposer) Start() error {
	numOurs := 0
	numTotal := 0
	for _, w := range p.chain.State().TipState().ValidatorRegistry {
		if _, ok := p.Keystore.GetValidatorKey(w.PubKey); ok {
			numOurs++
		}
		numTotal++
	}

	p.log.Infof("starting proposer with %d/%d active validators", numOurs, numTotal)

	go p.VoteForBlocks()
	go p.ProposeBlocks()

	return nil
}
