package proposer

import (
	"bytes"
	"context"
	"fmt"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
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
	hostnode       hostnode.HostNode
	blockTopic     *pubsub.Topic
	voteTopic      *pubsub.Topic

	lastActionManager actionmanager.LastActionManager
}

// NewProposer creates a new proposer from the parameters.
func NewProposer(log logger.Logger, params *params.ChainParams, chain chain.Blockchain, hostnode hostnode.HostNode, voteMempool mempool.VoteMempool, coinsMempool mempool.CoinsMempool, actionsMempool mempool.ActionMempool, manager actionmanager.LastActionManager, ks keystore.Keystore) (Proposer, error) {
	ctx, cancel := context.WithCancel(context.Background())
	blockTopic, err := hostnode.Topic(p2p.MsgBlockCmd)
	if err != nil {
		cancel()
		return nil, err
	}
	voteTopic, err := hostnode.Topic(p2p.MsgVoteCmd)
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
			return prop, nil
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
	msg := &p2p.MsgVote{Data: v}
	buf := bytes.NewBuffer([]byte{})
	err := p2p.WriteMessage(buf, msg, p.hostnode.GetNetMagic())
	if err != nil {
		p.log.Errorf("error encoding vote: %s", err)
		return
	}
	if err := p.voteTopic.Publish(p.context, buf.Bytes()); err != nil {
		p.log.Errorf("error publishing vote: %s", err)
	}
}

func (p *proposer) publishBlock(block *primitives.Block) {
	msg := &p2p.MsgBlock{Data: block}
	buf := bytes.NewBuffer([]byte{})
	err := p2p.WriteMessage(buf, msg, p.hostnode.GetNetMagic())
	if err != nil {
		p.log.Errorf("error encoding vote: %s", err)
		return
	}
	if err := p.blockTopic.Publish(p.context, buf.Bytes()); err != nil {
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

			// Check if we're an attester for this slot
			if p.hostnode.PeersConnected() == 0 || p.hostnode.Syncing() {
				blockTimer = time.NewTimer(time.Second * 10)
				p.log.Info("blockchain not synced... trying to propose in 10 seconds")
				continue
			}

			tip := p.chain.State().Tip()
			tipHash := tip.Hash

			voteState, err := p.chain.State().TipStateAtSlot(slotToPropose)
			if err != nil {
				p.log.Error("unable to get tip state at slot %d", slotToPropose)
				blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
				continue
			}

			slotIndex := (slotToPropose + p.params.EpochLength - 1) % p.params.EpochLength
			proposerIndex := voteState.GetProposerQueue()[slotIndex]
			proposer := voteState.GetValidatorRegistry()[proposerIndex]

			if k, found := p.keystore.GetValidatorKey(proposer.PubKey); found {

				//if !p.lastActionManager.ShouldRun(proposer.PubKey) {
				//	blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
				//	continue
				//}

				p.log.Infof("proposing for slot %d", slotToPropose)

				votes, err := p.voteMempool.Get(slotToPropose, voteState, p.params, proposerIndex)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
					continue
				}

				depositTxs, voteState, err := p.actionsMempool.GetDeposits(int(p.params.MaxDepositsPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
					continue
				}

				coinTxs, voteState := p.coinsMempool.Get(p.params.MaxTxsPerBlock, voteState)

				coinTxMulti := p.coinsMempool.GetMulti(p.params.MaxTxsMultiPerBlock, voteState)

				exitTxs, err := p.actionsMempool.GetExits(int(p.params.MaxExitsPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
					continue
				}

				randaoSlashings, err := p.actionsMempool.GetRANDAOSlashings(int(p.params.MaxRANDAOSlashingsPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
					continue
				}

				voteSlashings, err := p.actionsMempool.GetVoteSlashings(int(p.params.MaxVoteSlashingsPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
					continue
				}

				proposerSlashings, err := p.actionsMempool.GetProposerSlashings(int(p.params.MaxProposerSlashingsPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
					continue
				}

				governanceVotes, err := p.actionsMempool.GetGovernanceVotes(int(p.params.MaxGovernanceVotesPerBlock), voteState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
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
				block.Header.TxMultiMerkleRoot = block.TransactionMultiMerkleRoot()
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
					blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
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

			// Check if we're an attester for this slot
			if p.hostnode.PeersConnected() == 0 || p.hostnode.Syncing() {
				voteTimer = time.NewTimer(time.Second * 10)
				p.log.Info("blockchain not synced... trying to vote in 10 seconds")
				continue
			}

			s := p.chain.State()

			voteState, err := s.TipStateAtSlot(slotToVote)
			if err != nil {
				p.log.Errorf("unable to get tip at slot %d", slotToVote)
				voteTimer = time.NewTimer(time.Until(p.getNextVoteTime(slotToVote)))
				continue
			}

			validators, err := voteState.GetVoteCommittee(slotToVote, p.params)
			if err != nil {
				p.log.Errorf("error getting vote committee: %s", err.Error())
				voteTimer = time.NewTimer(time.Until(p.getNextVoteTime(slotToVote)))
				continue
			}

			p.log.Debugf("committing for slot %d with %d validators", slotToVote, len(validators))

			toEpoch := (slotToVote - 1) / p.params.EpochLength

			beaconBlock, found := s.Chain().GetNodeBySlot(slotToVote - 1)
			if !found {
				p.log.Errorf("unable to find block at slot %d", slotToVote-1)
				voteTimer = time.NewTimer(time.Until(p.getNextVoteTime(slotToVote)))
				continue
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

			var signatures []*bls.Signature

			bitlistVotes := bitfield.NewBitlist(uint64(len(validators)))

			validatorRegistry := voteState.GetValidatorRegistry()
			for i, index := range validators {
				votingValidator := validatorRegistry[index]
				key, found := p.keystore.GetValidatorKey(votingValidator.PubKey)
				if !found {
					continue
				}
				//signFunc := func(message *primitives.ValidatorHelloMessage) *bls.Signature {
				//	msg := message.SignatureMessage()
				//	return key.Sign(msg)
				//}
				//if p.lastActionManager.StartValidator(votingValidator.PubKey, signFunc) {
				signatures = append(signatures, key.Sign(dataHash[:]))
				bitlistVotes.Set(uint(i))
				//}
			}

			if len(signatures) > 0 {
				sig := bls.AggregateSignatures(signatures)

				var voteSig [96]byte
				copy(voteSig[:], sig.Marshal())

				vote := &primitives.MultiValidatorVote{
					Data:                  data,
					ParticipationBitfield: bitlistVotes,
					Sig:                   voteSig,
				}

				p.voteMempool.Add(vote)

				p.log.Infof("sending votes for slot %d for %d validators", slotToVote, len(signatures))

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
	go p.StartRoutine()
	return nil
}

func (p *proposer) Stop() {
	p.chain.Unnotify(p)
	p.stop()
}

func (p *proposer) Keystore() keystore.Keystore {
	return p.keystore
}

// The StartRoutine is a concurrent process that checks if the node should be voting/proposing
// It also monitors the vote and propose routines to prevent the node to stop doing those process.
func (p *proposer) StartRoutine() {

check:
	numOurs := 0
	numTotal := 0
	for _, w := range p.chain.State().TipState().GetValidatorRegistry() {
		_, ok := p.keystore.GetValidatorKey(w.PubKey)
		if ok {
			numOurs++
		}
		numTotal++
	}

	if numOurs == 0 {
		p.log.Info("there are no validators to vote/propose, retrying in 1 seconds")
		time.Sleep(time.Second * 10)
		goto check
	}

	p.log.Infof("starting proposer with %d/%d active validators", numOurs, numTotal)

	go p.VoteForBlocks()
	go p.ProposeBlocks()

}
