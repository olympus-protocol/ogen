package proposer

import (
	"context"
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"sync"
	"time"

	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/host"
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
	log       logger.Logger
	netParams *params.ChainParams
	chain     chain.Blockchain
	keystore  keystore.Keystore

	context context.Context
	stop    context.CancelFunc

	proposerLock sync.Mutex
	voteLock     sync.Mutex

	voting    bool
	proposing bool

	pool mempool.Pool
	host host.Host

	lastActionManager actionmanager.LastActionManager
}

// NewProposer creates a new proposer from the parameters.
func NewProposer(chain chain.Blockchain, h host.Host, pool mempool.Pool, ks keystore.Keystore, manager actionmanager.LastActionManager) (Proposer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	prop := &proposer{
		log:               config.GlobalParams.Logger,
		netParams:         config.GlobalParams.NetParams,
		keystore:          ks,
		chain:             chain,
		context:           ctx,
		stop:              cancel,
		pool:              pool,
		host:              h,
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
	p.pool.RemoveByBlock(block, newState)
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
func (p *proposer) ProposerSlashingConditionViolated(d *primitives.ProposerSlashing) {
	p.log.Warn("WARNING: Proposer slashing condition detected.")
	err := p.pool.AddProposerSlashing(d)
	if err != nil {
		p.log.Error(err)
	}
}

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
			if !p.host.Synced() {
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

				ok, err := p.lastActionManager.ShouldRun(proposerValidator.PubKey)
				if err != nil {
					blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
					p.log.Error(err)
					continue
				}
				if !ok {
					blockTimer = time.NewTimer(time.Until(p.getNextBlockTime(slotToPropose)))
					p.log.Info("proposing disable, another node is already proposing for this key")
					continue
				}

				p.log.Infof("proposing for slot %d", slotToPropose)

				votes := p.pool.GetVotes(slotToPropose, blockState, proposerIndex)

				deposits, blockState := p.pool.GetDeposits(blockState)

				exits, blockState := p.pool.GetExits(blockState)

				partialExits, blockState := p.pool.GetPartialExits(blockState)

				coinProofs, blockState := p.pool.GetCoinProofs(blockState)

				txs, blockState := p.pool.GetTxs(blockState, proposerValidator.PayeeAddress)

				voteSlashings, blockState := p.pool.GetVoteSlashings(blockState)

				proposerSlashings, blockState := p.pool.GetProposerSlashings(blockState)

				randaoSlashings, blockState := p.pool.GetRANDAOSlashings(blockState)

				governanceVotes, blockState := p.pool.GetGovernanceVotes(blockState)

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
					Deposits:          deposits,
					Exits:             exits,
					PartialExit:       partialExits,
					CoinProofs:        coinProofs,
					Txs:               txs,
					VoteSlashings:     voteSlashings,
					ProposerSlashings: proposerSlashings,
					RANDAOSlashings:   randaoSlashings,
					GovernanceVotes:   governanceVotes,
				}

				block.Header.VoteMerkleRoot = block.VotesMerkleRoot()
				block.Header.DepositMerkleRoot = block.DepositMerkleRoot()
				block.Header.ExitMerkleRoot = block.ExitMerkleRoot()
				block.Header.PartialExitMerkleRoot = block.PartialExitsMerkleRoot()
				block.Header.CoinProofsMerkleRoot = block.CoinProofsMerkleRoot()
				block.Header.TxsMerkleRoot = block.TxsMerkleRoot()
				block.Header.VoteSlashingMerkleRoot = block.VoteSlashingRoot()
				block.Header.ProposerSlashingMerkleRoot = block.ProposerSlashingsRoot()
				block.Header.RANDAOSlashingMerkleRoot = block.RANDAOSlashingsRoot()
				block.Header.GovernanceVotesMerkleRoot = block.GovernanceVoteMerkleRoot()
				block.Header.MultiSignatureTxsMerkleRoot = block.MultiSignatureTxsMerkleRoot()

				blockHash := block.Hash()
				randaoHash := chainhash.HashH([]byte(fmt.Sprintf("%d", slotToPropose)))

				if !k.Enable {
					blockTimer = time.NewTimer(time.Second * 2)
					p.proposerLock.Unlock()
					continue
				}

				blockSig := k.Secret.Sign(blockHash[:])
				randaoSig := k.Secret.Sign(randaoHash[:])
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
			if !p.host.Synced() {
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

			var signatures []common.Signature

			bitlistVotes := bitfield.NewBitlist(uint64(len(validators)))

			validatorRegistry := voteState.GetValidatorRegistry()

			validatorsActionMap := make(map[common.PublicKey]common.SecretKey)

			for i, index := range validators {
				votingValidator := validatorRegistry[index]
				key, ok := p.keystore.GetValidatorKey(votingValidator.PubKey)
				if ok {
					if key.Enable {
						ok, err := p.lastActionManager.ShouldRun(votingValidator.PubKey)
						if err != nil {
							p.log.Error(err)
							continue
						}
						if ok {
							signatures = append(signatures, key.Secret.Sign(dataHash[:]))
							bitlistVotes.Set(uint(i))
							validatorsActionMap[key.Secret.PublicKey()] = key.Secret
						}
					}
				}
			}

			if len(validatorsActionMap) > 0 {
				err = p.lastActionManager.StartValidators(validatorsActionMap)
				if err != nil {
					p.log.Error(err)
					voteTimer = time.NewTimer(time.Second * 2)
					p.voteLock.Unlock()
					continue
				}
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

				err = p.pool.AddVote(vote, voteState)
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
		p.log.Info("there are no validators to vote/propose, retrying in 10 seconds")
		time.Sleep(time.Second * 10)
		goto check
	}

	p.log.Infof("starting proposer with %d/%d active validators", numOurs, numTotal)

	go p.VoteForBlocks()
	go p.ProposeBlocks()

}
