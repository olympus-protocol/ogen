package proposer

import (
	"context"
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"sync"
	"time"

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
	GetCurrentSlot() uint64
	Voting() bool
	Proposing() bool
	Keystore() keystore.Keystore
}

var _ Proposer = &proposer{}

// proposer manages mining for the blockchain.
type proposer struct {
	log        logger.Logger
	netParams  *params.ChainParams
	chain      chain.Blockchain
	keystore   keystore.Keystore
	mineActive bool
	context    context.Context
	stop       context.CancelFunc

	proposerLock sync.Mutex
	voteLock     sync.Mutex

	voting    bool
	proposing bool

	voteMempool    mempool.VoteMempool
	coinsMempool   mempool.CoinsMempool
	actionsMempool mempool.ActionMempool
	host           hostnode.HostNode

	lastActionManager actionmanager.LastActionManager
}

// NewProposer creates a new proposer from the parameters.
func NewProposer(chain chain.Blockchain, hostnode hostnode.HostNode, voteMempool mempool.VoteMempool, coinsMempool mempool.CoinsMempool, actionsMempool mempool.ActionMempool, manager actionmanager.LastActionManager, ks keystore.Keystore) (Proposer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	prop := &proposer{
		log:               config.GlobalParams.Logger,
		netParams:         config.GlobalParams.NetParams,
		keystore:          ks,
		chain:             chain,
		mineActive:        true,
		context:           ctx,
		stop:              cancel,
		voteMempool:       voteMempool,
		coinsMempool:      coinsMempool,
		actionsMempool:    actionsMempool,
		host:              hostnode,
		lastActionManager: manager,
		voting:            false,
		proposing:         false,
	}

	err := prop.keystore.OpenKeystore()
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

func (p *proposer) Voting() bool {
	return p.voting
}

func (p *proposer) Proposing() bool {
	return p.proposing
}

// NewTip implements the BlockchainNotifee interface.
func (p *proposer) NewTip(_ *chainindex.BlockRow, block *primitives.Block, newState state.State, _ []*primitives.EpochReceipt) {
	p.voteMempool.Remove(block)
	p.coinsMempool.RemoveByBlock(block)
	p.actionsMempool.RemoveByBlock(block, newState)
}

func (p *proposer) GetCurrentSlot() uint64 {
	return p.getCurrentSlot()
}

func (p *proposer) getCurrentSlot() uint64 {
	slot := time.Now().Sub(p.chain.GenesisTime()) / (time.Duration(p.netParams.SlotDuration) * time.Second)
	if slot < 0 {
		return 0
	}
	return uint64(slot)
}

// getNextSlotTime gets the next slot time.
func (p *proposer) getNextBlockTime(nextSlot uint64) time.Time {
	return p.chain.GenesisTime().Add(time.Duration(nextSlot*p.netParams.SlotDuration) * time.Second)
}

// getNextSlotTime gets the next slot time.
func (p *proposer) getNextVoteTime(nextSlot uint64) time.Time {
	return p.chain.GenesisTime().Add(time.Duration(nextSlot*p.netParams.SlotDuration) * time.Second).Add(-time.Second * time.Duration(p.netParams.SlotDuration) / 2)
}

// ProposerSlashingConditionViolated implements chain notifee.
func (p *proposer) ProposerSlashingConditionViolated(_ *primitives.ProposerSlashing) {}

func (p *proposer) ProposeBlocks() {
	defer func() {
		p.proposing = false
	}()

	slotToPropose := p.getCurrentSlot() + 1

	blockTimer := time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))

	for {
		select {
		case <-blockTimer.C:
			p.proposerLock.Lock()
			// Check if we're an attester for this slot
			if p.host.Syncing() {
				p.proposing = false
				blockTimer = time.NewTimer(time.Second * 10)
				p.log.Info("blockchain not synced... trying to propose in 10 seconds")
				p.proposerLock.Unlock()
				continue
			}
			p.proposing = true

			tip := p.chain.State().Tip()
			tipHash := tip.Hash

			blockState, err := p.chain.State().TipStateAtSlot(slotToPropose)
			if err != nil {
				p.log.Errorf("unable to get tip state at slot %d", slotToPropose)
				blockTimer = time.NewTimer(time.Second * 2)
				p.proposerLock.Unlock()
				continue
			}

			slotIndex := (slotToPropose + p.netParams.EpochLength - 1) % p.netParams.EpochLength
			proposerIndex := blockState.GetProposerQueue()[slotIndex]
			proposerValidator := blockState.GetValidatorRegistry()[proposerIndex]

			if k, found := p.keystore.GetValidatorKey(proposerValidator.PubKey); found {

				//if !p.lastActionManager.ShouldRun(proposer.PubKey) {
				//	blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
				//	continue
				//}

				p.log.Infof("proposing for slot %d", slotToPropose)

				votes, err := p.voteMempool.Get(slotToPropose, blockState, proposerIndex)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}

				depositTxs, blockState, err := p.actionsMempool.GetDeposits(int(p.netParams.MaxDepositsPerBlock), blockState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}

				coinTxs, blockState := p.coinsMempool.Get(p.netParams.MaxTxsPerBlock, blockState, proposerValidator.PayeeAddress)

				coinTxMulti := p.coinsMempool.GetMulti(p.netParams.MaxTxsMultiPerBlock, blockState)

				exitTxs, err := p.actionsMempool.GetExits(int(p.netParams.MaxExitsPerBlock), blockState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}

				randaoSlashings, err := p.actionsMempool.GetRANDAOSlashings(int(p.netParams.MaxRANDAOSlashingsPerBlock), blockState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}

				voteSlashings, err := p.actionsMempool.GetVoteSlashings(int(p.netParams.MaxVoteSlashingsPerBlock), blockState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}

				proposerSlashings, err := p.actionsMempool.GetProposerSlashings(int(p.netParams.MaxProposerSlashingsPerBlock), blockState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}

				governanceVotes, err := p.actionsMempool.GetGovernanceVotes(int(p.netParams.MaxGovernanceVotesPerBlock), blockState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}

				/*coinproofs, err := p.actionsMempool.GetProofs(int(p.netParams.MaxCoinProofsPerBlock), blockState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}*/

				/*partialExits, err := p.actionsMempool.GetPartialExits(int(p.netParams.MaxPartialExitsPerBlock), blockState)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}*/

				block := primitives.Block{
					Header: &primitives.BlockHeader{
						Version:       0,
						Nonce:         p.lastActionManager.GetNonce(),
						PrevBlockHash: tipHash,
						Timestamp:     uint64(time.Now().Unix()),
						Slot:          slotToPropose,
						FeeAddress:    proposerValidator.PayeeAddress,
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
					//CoinProofs:        coinproofs,
					//PartialExit:       partialExits,
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
				//block.Header.CoinProofsMerkleRoot = block.CoinProofsMerkleRoot()
				//block.Header.PartialExitMerkleRoot = block.PartialExitsMerkleRoot()

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
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}

				msg := &p2p.MsgBlock{Data: &block}

				err = p.host.Broadcast(msg)
				if err != nil {
					p.log.Error(err)
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}
			}

			slotToPropose++
			p.proposerLock.Unlock()
			blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
		case <-p.context.Done():
			p.log.Info("stopping proposer")
			return
		}
	}
}

func (p *proposer) VoteForBlocks() {
	defer func() {
		p.voting = false
	}()

	slotToVote := p.getCurrentSlot() + 1
	if slotToVote <= 0 {
		slotToVote = 1
	}

	voteTimer := time.NewTimer(time.Until(p.getNextVoteTime(slotToVote)))

	for {
		select {
		case <-voteTimer.C:
			p.voteLock.Lock()
			// Check if we're an attester for this slot
			if p.host.Syncing() {
				p.voting = false
				voteTimer = time.NewTimer(time.Second * 10)
				p.log.Info("blockchain not synced... trying to vote in 10 seconds")
				p.voteLock.Unlock()
				continue
			}

			p.voting = true
			s := p.chain.State()

			voteState, err := s.TipStateAtSlot(slotToVote)
			if err != nil {
				p.log.Errorf("unable to get tip at slot %d", slotToVote)
				voteTimer = time.NewTimer(time.Second * 2)
				p.voteLock.Unlock()
				continue
			}

			validators, err := voteState.GetVoteCommittee(slotToVote)
			if err != nil {
				p.log.Errorf("error getting vote committee: %s", err.Error())
				voteTimer = time.NewTimer(time.Second * 2)
				p.voteLock.Unlock()
				continue
			}

			p.log.Debugf("committing for slot %d with %d validators", slotToVote, len(validators))

			beaconBlock, found := s.Chain().GetNodeBySlot(slotToVote - 1)
			if !found {
				p.log.Errorf("unable to find block at slot %d", slotToVote-1)
				voteTimer = time.NewTimer(time.Second * 2)
				p.voteLock.Unlock()
				continue
			}

			toEpoch := (slotToVote - 1) / p.netParams.EpochLength

			data := &primitives.VoteData{
				Slot:            slotToVote,
				FromEpoch:       voteState.GetJustifiedEpoch(),
				FromHash:        voteState.GetJustifiedEpochHash(),
				ToEpoch:         toEpoch,
				ToHash:          voteState.GetRecentBlockHash(toEpoch*p.netParams.EpochLength - 1),
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

				err = p.voteMempool.AddValidate(vote, voteState)
				if err != nil {
					p.log.Error(err)
					voteTimer = time.NewTimer(time.Second * 2)
					p.voteLock.Unlock()
					continue
				}

				p.log.Infof("sending votes for slot %d for %d validators", slotToVote, len(signatures))

				msg := &p2p.MsgVote{Data: vote}
				err = p.host.Broadcast(msg)
				if err != nil {
					p.log.Error(err)
					voteTimer = time.NewTimer(time.Second * 2)
					p.voteLock.Unlock()
					continue
				}
			}

			slotToVote++
			p.voteLock.Unlock()
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
