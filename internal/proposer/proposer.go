package proposer

import (
	"context"
	"fmt"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/peers"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// Config is a config for the proposer.
type Config struct {
	Datadir string
	Log     logger.Logger
}

// proposer manages mining for the blockchain.
type Proposer struct {
	log        logger.Logger
	config     Config
	params     params.ChainParams
	chain      chain.Blockchain
	Keystore   *keystore.Keystore
	mineActive bool
	context    context.Context
	Stop       context.CancelFunc

	voteMempool    mempool.VoteMempool
	coinsMempool   *mempool.CoinsMempool
	actionsMempool *mempool.ActionMempool
	hostnode       peers.HostNode
	blockTopic     *pubsub.Topic
	voteTopic      *pubsub.Topic

	lastActionManager actionmanager.LastActionManager
}

// OpenKeystore opens the keystore with the provided password returns error if the keystore doesn't exist.
func (p *Proposer) OpenKeystore() (err error) {
	p.Keystore = keystore.NewKeystore(p.config.Datadir, p.log)
	err = p.Keystore.OpenKeystore()
	if err != nil {
		return err
	}
	return nil
}

// NewProposer creates a new proposer from the parameters.
func NewProposer(config Config, params params.ChainParams, chain chain.Blockchain, hostnode peers.HostNode, voteMempool mempool.VoteMempool, coinsMempool *mempool.CoinsMempool, actionsMempool *mempool.ActionMempool, manager actionmanager.LastActionManager) (proposer *Proposer, err error) {
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
		log:               config.Log,
		config:            config,
		params:            params,
		chain:             chain,
		mineActive:        true,
		context:           ctx,
		Stop:              cancel,
		voteMempool:       voteMempool,
		coinsMempool:      coinsMempool,
		actionsMempool:    actionsMempool,
		hostnode:          hostnode,
		blockTopic:        blockTopic,
		voteTopic:         voteTopic,
		lastActionManager: manager,
	}
	chain.Notify(proposer)
	return proposer, nil
}

// NewTip implements the BlockchainNotifee interface.
func (p *Proposer) NewTip(_ *chainindex.BlockRow, block *primitives.Block, newState state.State, _ []*primitives.EpochReceipt) {
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

func (p *Proposer) publishVotes(v *primitives.MultiValidatorVote) {
	buf, err := v.Marshal()
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
			if p.hostnode.PeersConnected() == 0 || p.hostnode.Syncing() {
				p.log.Infof("blockchain not synced... trying to mine in 10 seconds")
				blockTimer = time.NewTimer(time.Second * 10)
				continue
			}

			// check if we're an attester for this slot
			tip := p.chain.State().Tip()
			tipHash := tip.Hash

			state, err := p.chain.State().TipStateAtSlot(slotToPropose)
			if err != nil {
				p.log.Error(err)
				continue
			}

			slotIndex := (slotToPropose + p.params.EpochLength - 1) % p.params.EpochLength
			proposerIndex := state.GetProposerQueue()[slotIndex]
			proposer := state.GetValidatorRegistry()[proposerIndex]

			if k, found := p.Keystore.GetValidatorKey(proposer.PubKey); found {

				if !p.lastActionManager.ShouldRun(proposer.PubKey) {
					continue
				}

				p.log.Infof("proposing for slot %d", slotToPropose)

				votes, err := p.voteMempool.Get(slotToPropose, state, &p.params, proposerIndex)
				if err != nil {
					p.log.Error(err)
					continue
				}

				depositTxs, state, err := p.actionsMempool.GetDeposits(int(p.params.MaxDepositsPerBlock), state)
				if err != nil {
					p.log.Error(err)
					continue
				}

				coinTxs, state := p.coinsMempool.Get(p.params.MaxTxsPerBlock, state)

				exitTxs, err := p.actionsMempool.GetExits(int(p.params.MaxExitsPerBlock), state)
				if err != nil {
					p.log.Error(err)
					continue
				}

				randaoSlashings, err := p.actionsMempool.GetRANDAOSlashings(int(p.params.MaxRANDAOSlashingsPerBlock), state)
				if err != nil {
					p.log.Error(err)
					continue
				}

				voteSlashings, err := p.actionsMempool.GetVoteSlashings(int(p.params.MaxVoteSlashingsPerBlock), state)
				if err != nil {
					p.log.Error(err)
					continue
				}

				proposerSlashings, err := p.actionsMempool.GetProposerSlashings(int(p.params.MaxProposerSlashingsPerBlock), state)
				if err != nil {
					p.log.Error(err)
					return
				}

				governanceVotes, err := p.actionsMempool.GetGovernanceVotes(int(p.params.MaxGovernanceVotesPerBlock), state)
				if err != nil {
					p.log.Error(err)
					continue
				}

				block := primitives.Block{
					Header: &primitives.BlockHeader{
						Version:       0,
						Nonce:         p.lastActionManager.GetNonce(),
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
					GovernanceVotes:   governanceVotes,
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
					continue
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
			if p.hostnode.PeersConnected() == 0 || p.hostnode.Syncing() {
				voteTimer = time.NewTimer(time.Second * 10)
				p.log.Infof("blockchain not synced... trying to mine in 10 seconds")
				continue
			}

			s := p.chain.State()

			state, err := s.TipStateAtSlot(slotToVote)
			if err != nil {
				panic(err)
			}

			validators, err := state.GetVoteCommittee(slotToVote, &p.params)
			if err != nil {
				p.log.Errorf("error getting vote committee: %s", err.Error())
				continue
			}

			p.log.Debugf("committing for slot %d with %d validators", slotToVote, len(validators))

			toEpoch := (slotToVote - 1) / p.params.EpochLength

			beaconBlock, found := s.Chain().GetNodeBySlot(slotToVote - 1)
			if !found {
				panic("could not find block")
			}

			data := &primitives.VoteData{
				Slot:            slotToVote,
				FromEpoch:       state.GetJustifiedEpoch(),
				FromHash:        state.GetJustifiedEpochHash(),
				ToEpoch:         toEpoch,
				ToHash:          state.GetRecentBlockHash(toEpoch*p.params.EpochLength-1, &p.params),
				BeaconBlockHash: beaconBlock.Hash,
				Nonce:           p.lastActionManager.GetNonce(),
			}

			dataHash := data.Hash()

			vote := &primitives.MultiValidatorVote{
				Data:                  data,
				ParticipationBitfield: bitfield.NewBitlist(uint64(len(validators))),
			}

			var signatures []*bls.Signature

			for i, validatorIdx := range validators {
				validator := state.GetValidatorRegistry()[validatorIdx]
				if k, found := p.Keystore.GetValidatorKey(validator.PubKey); found {
					if !p.lastActionManager.ShouldRun(validator.PubKey) {
						return
					}
					signatures = append(signatures, k.Sign(dataHash[:]))

					vote.ParticipationBitfield.Set(uint(i))
				}
			}

			if len(signatures) > 0 {
				sig := bls.AggregateSignatures(signatures)

				var voteSig [96]byte
				copy(voteSig[:], sig.Marshal())
				vote.Sig = voteSig

				err = p.voteMempool.AddValidate(vote, state)
				if err != nil {
					p.log.Error("unable to submit own generated vote")
					return
				}

				go p.publishVotes(vote)
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
	for _, w := range p.chain.State().TipState().GetValidatorRegistry() {
		secKey, ok := p.Keystore.GetValidatorKey(w.PubKey)
		if ok {
			numOurs++
		}
		if ok {
			p.lastActionManager.StartValidator(w.PubKey, func(message *actionmanager.ValidatorHelloMessage) *bls.Signature {
				msg := message.SignatureMessage()
				return secKey.Sign(msg)
			})
		}
		numTotal++
	}

	p.log.Infof("starting proposer with %d/%d active validators", numOurs, numTotal)

	go p.VoteForBlocks()
	go p.ProposeBlocks()

	return nil
}
