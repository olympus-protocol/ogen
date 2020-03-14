package chain

import (
	"errors"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/state"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"sync"
)

type Step uint8

const (
	StepPropose Step = iota
	StepPrevote
	StepPrecommit
)

type consensusVoteIndex struct {
	blockHash chainhash.Hash
	round     uint64
}

type consensusVoteData struct {
	voted map[chainhash.Hash]bool
	num   uint64
}

type messageLog struct {
	proposals       []p2p.MsgProposal
	prevotes        map[consensusVoteIndex]*consensusVoteData
	precommits      map[consensusVoteIndex]*consensusVoteData
	prevoteRounds   map[uint64]uint64
	precommitRounds map[uint64]uint64
}

func newMessageLog() messageLog {
	return messageLog{
		prevotes:        map[consensusVoteIndex]*consensusVoteData{},
		precommits:      map[consensusVoteIndex]*consensusVoteData{},
		prevoteRounds:   map[uint64]uint64{},
		precommitRounds: map[uint64]uint64{},
	}
}

func (m *messageLog) numVotes(h chainhash.Hash, round uint64) uint64 {
	numVotes, found := m.prevotes[consensusVoteIndex{blockHash: h, round: round}]
	if !found {
		return 0
	}

	return numVotes.num
}

func (m *messageLog) hasVoted(h chainhash.Hash, round uint64, validator chainhash.Hash) bool {
	votes, found := m.prevotes[consensusVoteIndex{
		blockHash: h,
		round:     round,
	}]
	if !found {
		return false
	}

	_, found = votes.voted[validator]
	return found
}

func (m *messageLog) numCommits(h chainhash.Hash, round uint64) uint64 {
	numVotes, found := m.precommits[consensusVoteIndex{blockHash: h, round: round}]
	if !found {
		return 0
	}

	return numVotes.num
}

func (m *messageLog) hasCommitted(h chainhash.Hash, round uint64, validator chainhash.Hash) bool {
	votes, found := m.precommits[consensusVoteIndex{
		blockHash: h,
		round:     round,
	}]
	if !found {
		return false
	}

	_, found = votes.voted[validator]
	return found
}

type ProposalPredicate = func(proposal p2p.MsgProposal) bool

func (m *messageLog) getProposals(pred ProposalPredicate) []p2p.MsgProposal {
	proposals := make([]p2p.MsgProposal, 0, len(m.proposals))
	for _, p := range m.proposals {
		if pred(p) {
			proposals = append(proposals, p)
		}
	}
	return proposals
}

func (m *messageLog) receiveProposal(proposal p2p.MsgProposal) {
	m.proposals = append(m.proposals, proposal)
}

func (m *messageLog) receiveVote(vote p2p.MsgPrevote) {
	if m.hasVoted(vote.BlockHash, vote.Round, vote.ValidatorID) {
		return
	}

	_, found := m.prevotes[consensusVoteIndex{blockHash: vote.BlockHash, round: vote.Round}]
	if !found {
		m.prevotes[consensusVoteIndex{blockHash: vote.BlockHash, round: vote.Round}] = &consensusVoteData{
			num: 1,
			voted: map[chainhash.Hash]bool{
				vote.ValidatorID: true,
			},
		}
	} else {
		m.prevotes[consensusVoteIndex{blockHash: vote.BlockHash, round: vote.Round}].num++
		m.prevotes[consensusVoteIndex{blockHash: vote.BlockHash, round: vote.Round}].voted[vote.ValidatorID] = true
	}

	if _, found := m.prevoteRounds[vote.Round]; found {
		m.prevoteRounds[vote.Round]++
	} else {
		m.prevoteRounds[vote.Round] = 1
	}
}

func (m *messageLog) receiveCommit(vote p2p.MsgPrecommit) {
	if m.hasCommitted(vote.BlockHash, vote.Round, vote.ValidatorID) {
		return
	}

	_, found := m.precommits[consensusVoteIndex{blockHash: vote.BlockHash, round: vote.Round}]
	if !found {
		m.precommits[consensusVoteIndex{blockHash: vote.BlockHash, round: vote.Round}] = &consensusVoteData{
			num: 1,
			voted: map[chainhash.Hash]bool{
				vote.ValidatorID: true,
			},
		}
	} else {
		m.precommits[consensusVoteIndex{blockHash: vote.BlockHash, round: vote.Round}].num++
		m.precommits[consensusVoteIndex{blockHash: vote.BlockHash, round: vote.Round}].voted[vote.ValidatorID] = true
	}

	if _, found := m.precommitRounds[vote.Round]; found {
		m.precommitRounds[vote.Round]++
	} else {
		m.precommitRounds[vote.Round] = 1
	}
}

type Consensus struct {
	// Tendermint
	height      uint64
	round       uint64
	step        Step
	lockedValue chainhash.Hash
	lockedRound int64
	validValue  *primitives.Block
	validRound  int64

	messageLog messageLog

	chain ConsensusChainInterface
	miner ConsensusMinerInterface
	p2p   ConsensusP2PInterface
	lock  sync.Mutex
}

type ConsensusChainInterface interface {
	Valid(block primitives.Block) error
	GetValue() (*primitives.Block, error)
	GetProposer(round uint64) chainhash.Hash
	GetWorkerData(chainhash.Hash) (state.Worker, bool)
	NumWorkers() uint64
	Decide(chainhash.Hash)
}

type ConsensusMinerInterface interface {
	Sign(hash chainhash.Hash) (bls.Signature, error)
	ValidatorID() chainhash.Hash
}

type ConsensusP2PInterface interface {
	Broadcast(message p2p.Message)
}

// NewConsensus creates a new consensus instance.
func NewConsensus(chain ConsensusChainInterface, miner ConsensusMinerInterface, p2p ConsensusP2PInterface) (*Consensus, error) {
	c := &Consensus{
		height:     0,
		round:      0,
		step:       StepPropose,
		messageLog: newMessageLog(),
		chain:      chain,
		miner:      miner,
		p2p:        p2p,
		validRound: -1,
		lockedRound: -1,
	}

	return c, nil
}

func (c *Consensus) Initialize() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.startRound(0)
}

func (c *Consensus) startRound(round uint64) error {
	c.round = round
	c.step = StepPropose

	proposerID := c.chain.GetProposer(c.round)

	// Line 14
	validatorID := c.miner.ValidatorID()
	if c.miner != nil && validatorID.IsEqual(&proposerID) {
		var proposalBlock primitives.Block

		// Line 15-18
		if c.validValue != nil {
			proposalBlock = *c.validValue
		} else {
			p, err := c.chain.GetValue()
			if err != nil {
				return err
			}
			proposalBlock = *p
		}

		// Line 19
		proposal := &p2p.MsgProposal{
			Height:        c.height,
			Round:         c.round,
			BlockProposal: proposalBlock,
			ValidRound:    c.validRound,
			Signature:     [96]byte{},
			ValidatorID: c.miner.ValidatorID(),
		}

		sig, err := c.miner.Sign(proposal.Hash())
		if err != nil {
			return err
		}

		proposal.Signature = sig.Serialize()

		c.p2p.Broadcast(proposal)
	} else {
		// TODO: timeout
	}

	return nil
}

var (
	ErrSignatureInvalid = errors.New("signature is invalid")
	ErrMissingProposer  = errors.New("missing data for proposer")
)

func (c *Consensus) checkProposer(proposal p2p.MsgProposal) error {
	// make sure it came from the correct proposer
	proposerID := c.chain.GetProposer(c.round)

	proposer, ok := c.chain.GetWorkerData(proposerID)
	if !ok {
		return ErrMissingProposer
	}

	workerPub, err := bls.DeserializePublicKey(proposer.PubKey)
	if err != nil {
		return err
	}

	proposalSignature, err := bls.DeserializeSignature(proposal.Signature)
	if err != nil {
		return err
	}
	proposalHash := proposal.Hash()

	valid, err := bls.VerifySig(workerPub, proposalHash[:], proposalSignature)
	if err != nil {
		return err
	}

	if !valid {
		return ErrSignatureInvalid
	}

	return nil
}

func (c *Consensus) OnMessageProposal(proposal p2p.MsgProposal) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// TODO: do we need more verification here?

	// check height is current height
	if proposal.Height != c.height {
		// ignore at wrong height
		return nil
	}

	// check proposer is valid at current round
	if err := c.checkProposer(proposal); err != nil {
		return err
	}

	c.messageLog.receiveProposal(proposal)

	return c.checkRules()
}

func (c *Consensus) OnMessagePrevote(prevote p2p.MsgPrevote) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// TODO: do we need more verification here?
	if prevote.Height != c.height {
		// ignore at wrong height
		return nil
	}

	c.messageLog.receiveVote(prevote)

	return c.checkRules()
}

func (c *Consensus) OnMessagePrecommit(precommit p2p.MsgPrecommit) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// TODO: do we need more verification here?
	if precommit.Height != c.height {
		// ignore at wrong height
		return nil
	}

	c.messageLog.receiveCommit(precommit)

	return c.checkRules()
}

func withRoundAndValidRound(round uint64, validRound int64) ProposalPredicate {
	return func(proposal p2p.MsgProposal) bool {
		return proposal.Round == round && proposal.ValidRound == validRound
	}
}

func withRound(round uint64) ProposalPredicate {
	return func(proposal p2p.MsgProposal) bool {
		return proposal.Round == round
	}
}

func (c *Consensus) checkRules() error {
	if c.step == StepPropose {
		// Rule line 22
		matchingProposals := c.messageLog.getProposals(withRoundAndValidRound(c.round, -1))
		if len(matchingProposals) > 0 {
			proposal := matchingProposals[0]
			blockHash := proposal.BlockProposal.Hash()
			if c.miner != nil {
				if c.chain.Valid(proposal.BlockProposal) == nil && (c.lockedRound == -1 || c.lockedValue.IsEqual(&blockHash)) {
					vote := &p2p.MsgPrevote{
						Height:      c.height,
						Round:       c.round,
						BlockHash:   blockHash,
						ValidatorID: c.miner.ValidatorID(),
						Signature:   [96]byte{},
					}

					sig, err := c.miner.Sign(vote.Hash())
					if err != nil {
						return err
					}

					vote.Signature = sig.Serialize()

					c.p2p.Broadcast(vote)
				} else {
					vote := &p2p.MsgPrevote{
						Height:      c.height,
						Round:       c.round,
						ValidatorID: c.miner.ValidatorID(),
						Signature:   [96]byte{},
					}

					sig, err := c.miner.Sign(vote.Hash())
					if err != nil {
						return err
					}

					vote.Signature = sig.Serialize()

					c.p2p.Broadcast(vote)
				}
			}

			c.step = StepPrevote
		}

		// Rule line 28
		matchingProposals = c.messageLog.getProposals(func(proposal p2p.MsgProposal) bool {
			return proposal.ValidRound >= 0 && proposal.ValidRound < int64(c.round)
		})

		for _, p := range matchingProposals {
			blockHash := p.BlockProposal.Hash()
			if c.messageLog.numVotes(blockHash, uint64(p.ValidRound))*3 > c.chain.NumWorkers()*2 {
				if c.miner != nil {
					if c.chain.Valid(p.BlockProposal) == nil && (c.lockedRound <= p.ValidRound || c.lockedValue.IsEqual(&blockHash)) {
						vote := &p2p.MsgPrevote{
							Height:      c.height,
							Round:       c.round,
							BlockHash:   blockHash,
							ValidatorID: c.miner.ValidatorID(),
							Signature:   [96]byte{},
						}

						sig, err := c.miner.Sign(vote.Hash())
						if err != nil {
							return err
						}

						vote.Signature = sig.Serialize()

						c.p2p.Broadcast(vote)
					} else {
						vote := &p2p.MsgPrevote{
							Height:      c.height,
							Round:       c.round,
							ValidatorID: c.miner.ValidatorID(),
							Signature:   [96]byte{},
						}

						sig, err := c.miner.Sign(vote.Hash())
						if err != nil {
							return err
						}

						vote.Signature = sig.Serialize()

						c.p2p.Broadcast(vote)
					}
				}

				c.step = StepPrevote
			}
		}
	} else {
		// Rule line 36
		// Assuming: proposers are correct and step >= prevote

		// validValue is only set in here, so if validValue is nil, we haven't run this rule before as
		// required by the spec
		if c.validValue == nil {
			matchingProposals := c.messageLog.getProposals(withRound(c.round))
			for _, p := range matchingProposals {
				blockHash := p.BlockProposal.Hash()
				numPrevotes := c.messageLog.numVotes(blockHash, c.round)
				if c.chain.Valid(p.BlockProposal) != nil {
					continue
				}
				if numPrevotes*3 > c.chain.NumWorkers()*2 {
					// rule is matched
					if c.step == StepPrevote {
						c.lockedValue = p.BlockProposal.Hash()
						c.lockedRound = int64(c.round)

						if c.miner != nil {
							vote := &p2p.MsgPrecommit{
								Height:      c.height,
								Round:       c.round,
								BlockHash:   blockHash,
								ValidatorID: c.miner.ValidatorID(),
								Signature:   [96]byte{},
							}

							sig, err := c.miner.Sign(vote.Hash())
							if err != nil {
								return err
							}

							vote.Signature = sig.Serialize()

							c.p2p.Broadcast(vote)
						}

						c.step = StepPrecommit
					}

					c.validRound = p.ValidRound
					c.validValue = &p.BlockProposal
					break
				}
			}
		}
	}

	// Rule line 44
	if c.step == StepPrevote {
		numPrevotes := c.messageLog.numVotes(chainhash.Hash{}, c.round)

		if numPrevotes*3 > c.chain.NumWorkers() {
			c.step = StepPrecommit

			if c.miner != nil {
				vote := &p2p.MsgPrecommit{
					Height:      c.height,
					Round:       c.round,
					ValidatorID: c.miner.ValidatorID(),
					Signature:   [96]byte{},
				}

				sig, err := c.miner.Sign(vote.Hash())
				if err != nil {
					return err
				}

				vote.Signature = sig.Serialize()

				c.p2p.Broadcast(vote)
			}
		}
	}

	// Rule line 49
	for _, p := range c.messageLog.proposals {
		blockHash := p.BlockProposal.Hash()
		numPrecommits := c.messageLog.numCommits(blockHash, p.Round)

		if numPrecommits*3 > c.chain.NumWorkers() * 2 {
			c.chain.Decide(blockHash)
			c.height++
			c.lockedRound = -1
			c.lockedValue = chainhash.Hash{}
			c.validValue = nil
			c.validRound = -1
			c.messageLog = newMessageLog()
			if err := c.startRound(0); err != nil {
				return err
			}
		}
	}

	// Rule line 55
	for num, round := range c.messageLog.precommitRounds {
		if round <= c.round {
			continue
		}
		if num*3 > c.chain.NumWorkers() {
			err := c.startRound(round)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
