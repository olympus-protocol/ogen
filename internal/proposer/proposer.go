package proposer

import (
	"context"
	"fmt"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
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

// Proposer is the interface for proposer
type Proposer interface {
	NewTip(_ *chainindex.BlockRow, block *primitives.Block, newState state.State, _ []*primitives.EpochReceipt)
	ProposerSlashingConditionViolated(_ *primitives.ProposerSlashing)
	ProposeBlocks()
	VoteForBlocks()
	Start() error
	Stop()
	Keystore() keystore.Keystore
}

var _ Proposer = &proposer{}

// proposer manages mining for the blockchain.
type proposer struct {
	log        logger.Logger
	params     *params.ChainParams
	chain      chain.Blockchain
	keystore   keystore.Keystore
	mineActive bool
	context    context.Context
	stop       context.CancelFunc

	voteMempool    mempool.VoteMempool
	coinsMempool   mempool.CoinsMempool
	actionsMempool mempool.ActionMempool
	hostnode       peers.HostNode
	blockTopic     *pubsub.Topic
	voteTopic      *pubsub.Topic

	lastActionManager actionmanager.LastActionManager
}

// NewProposer creates a new proposer from the parameters.
func NewProposer(log logger.Logger, params *params.ChainParams, chain chain.Blockchain, hostnode peers.HostNode, voteMempool mempool.VoteMempool, coinsMempool mempool.CoinsMempool, actionsMempool mempool.ActionMempool, manager actionmanager.LastActionManager, ks keystore.Keystore) (Proposer, error) {
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

	prop := &proposer{
		log:               log,
		params:            params,
		keystore:          ks,
		chain:             chain,
		mineActive:        true,
		context:           ctx,
		stop:              cancel,
		voteMempool:       voteMempool,
		coinsMempool:      coinsMempool,
		actionsMempool:    actionsMempool,
		hostnode:          hostnode,
		blockTopic:        blockTopic,
		voteTopic:         voteTopic,
		lastActionManager: manager,
	}

	err = prop.keystore.OpenKeystore()
	if err != nil {
		if err == keystore.ErrorNotInitialized {
			err = prop.keystore.CreateKeystore()
			if err != nil {
				return nil, err
			}
		}
		return nil, err
	}
	return prop, nil
}

// NewTip implements the BlockchainNotifee interface.
func (p *proposer) NewTip(_ *chainindex.BlockRow, block *primitives.Block, newState state.State, _ []*primitives.EpochReceipt) {
	p.voteMempool.Remove(block)
	p.coinsMempool.RemoveByBlock(block)
	p.actionsMempool.RemoveByBlock(block, newState)
}

func (p *proposer) getCurrentSlot() uint64 {
	slot := time.Now().Sub(p.chain.GenesisTime()) / (time.Duration(p.params.SlotDuration) * time.Second)
	if slot < 0 {
		return 0
	}
	return uint64(slot)
}

// getNextSlotTime gets the next slot time.
func (p *proposer) getNextBlockTime(nextSlot uint64) time.Time {
	return p.chain.GenesisTime().Add(time.Duration(nextSlot*p.params.SlotDuration) * time.Second)
}

// getNextSlotTime gets the next slot time.
func (p *proposer) getNextVoteTime(nextSlot uint64) time.Time {
	return p.chain.GenesisTime().Add(time.Duration(nextSlot*p.params.SlotDuration) * time.Second).Add(-time.Second * time.Duration(p.params.SlotDuration) / 2)
}

func (p *proposer) publishVotes(v *primitives.MultiValidatorVote) {
	buf, err := v.Marshal()
	if err != nil {
		p.log.Errorf("error encoding vote: %s", err)
		return
	}

	if err := p.voteTopic.Publish(p.context, buf); err != nil {
		p.log.Errorf("error publishing vote: %s", err)
	}
}

func (p *proposer) publishBlock(block *primitives.Block) {
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
func (p *proposer) ProposerSlashingConditionViolated(_ *primitives.ProposerSlashing) {}

func (p *proposer) ProposeBlocks() {
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

			voteState, err := p.chain.State().TipStateAtSlot(slotToPropose)
			if err != nil {
				p.log.Error(err)
				blockTimer = time.NewTimer(time.Second * 10)
				continue
			}

			slotIndex := (slotToPropose + p.params.EpochLength - 1) % p.params.EpochLength
			proposerIndex := voteState.GetProposerQueue()[slotIndex]
			proposer := voteState.GetValidatorRegistry()[proposerIndex]

			if k, found := p.keystore.GetValidatorKey(proposer.PubKey); found {

				if !p.lastActionManager.ShouldRun(proposer.PubKey) {
					blockTimer = time.NewTimer(time.Second * 10)
					continue
				}

				p.log.Infof("proposing for slot %d", slotToPropose)

				votes, err := p.voteMempool.Get(slotToPropose, voteState, p.params, proposerIndex)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 10)
					continue
				}

				depositTxs, voteState, err := p.actionsMempool.GetDeposits(int(p.params.MaxDepositsPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 10)
					continue
				}

				coinTxs, voteState := p.coinsMempool.Get(p.params.MaxTxsPerBlock, voteState)

				coinTxMulti := p.coinsMempool.GetMulti(p.params.MaxTxsMultiPerBlock, voteState)

				exitTxs, err := p.actionsMempool.GetExits(int(p.params.MaxExitsPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 10)
					continue
				}

				randaoSlashings, err := p.actionsMempool.GetRANDAOSlashings(int(p.params.MaxRANDAOSlashingsPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 10)
					continue
				}

				voteSlashings, err := p.actionsMempool.GetVoteSlashings(int(p.params.MaxVoteSlashingsPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 10)
					continue
				}

				proposerSlashings, err := p.actionsMempool.GetProposerSlashings(int(p.params.MaxProposerSlashingsPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 10)
					continue
				}

				governanceVotes, err := p.actionsMempool.GetGovernanceVotes(int(p.params.MaxGovernanceVotesPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 10)
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
					TxsMulti:          coinTxMulti,
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
					blockTimer = time.NewTimer(time.Second * 10)
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

func (p *proposer) VoteForBlocks() {
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

			voteState, err := s.TipStateAtSlot(slotToVote)
			if err != nil {
				panic(err)
			}

			validators, err := voteState.GetVoteCommittee(slotToVote, p.params)
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
				FromEpoch:       voteState.GetJustifiedEpoch(),
				FromHash:        voteState.GetJustifiedEpochHash(),
				ToEpoch:         toEpoch,
				ToHash:          voteState.GetRecentBlockHash(toEpoch*p.params.EpochLength-1, p.params),
				BeaconBlockHash: beaconBlock.Hash,
				Nonce:           p.lastActionManager.GetNonce(),
			}

			dataHash := data.Hash()

			var signatures []bls_interface.Signature

			bitlistVotes := bitfield.NewBitlist(uint64(len(validators)))

			for i, index := range validators {
				votingValidator := voteState.GetValidatorRegistry()[index]
				key, found := p.keystore.GetValidatorKey(votingValidator.PubKey)
				if !found {
					continue
				}
				//signFunc := func(message *actionmanager.ValidatorHelloMessage) bls_interface.Signature {
				//	msg := message.SignatureMessage()
				//	return key.Sign(msg)
				//}
				//if p.lastActionManager.StartValidator(votingValidator.PubKey, signFunc) {
				signatures = append(signatures, key.Sign(dataHash[:]))
				bitlistVotes.Set(uint(i))
				//}
			}
			if len(signatures) > 0 {
				sig := bls.CurrImplementation.AggregateSignatures(signatures)

				var voteSig [96]byte
				copy(voteSig[:], sig.Marshal())

				vote := &primitives.MultiValidatorVote{
					Data:                  data,
					ParticipationBitfield: bitlistVotes,
					Sig:                   voteSig,
				}

				err = p.voteMempool.AddValidate(vote, voteState)
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
func (p *proposer) Start() error {
	p.chain.Notify(p)

	numOurs := 0
	numTotal := 0
	for _, w := range p.chain.State().TipState().GetValidatorRegistry() {
		_, ok := p.keystore.GetValidatorKey(w.PubKey)
		if ok {
			numOurs++
		}
		numTotal++
	}

	p.log.Infof("starting proposer with %d/%d active validators", numOurs, numTotal)

	go p.VoteForBlocks()
	go p.ProposeBlocks()

	return nil
}

func (p *proposer) Stop() {
	p.chain.Unnotify(p)
	p.stop()
}

func (p *proposer) Keystore() keystore.Keystore {
	return p.keystore
}
