package mempool

import (
	"context"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"sync"
)

type Pool interface {
	Load()
	Start()
	Stop()

	AddVote(d *primitives.MultiValidatorVote) error
	AddDeposit(d *primitives.Deposit) error
	AddExit(d *primitives.Exit) error
	AddPartialExit(d *primitives.PartialExit) error
	AddTx(d *primitives.Tx) error
	AddMultiSignatureTx(d *primitives.MultiSignatureTx) error
	AddVoteSlashing(d *primitives.VoteSlashing) error
	AddProposerSlashing(d *primitives.ProposerSlashing) error
	AddRANDAOSlashing(d *primitives.RANDAOSlashing) error
	AddGovernanceVote(d *primitives.GovernanceVote) error
	AddCoinProof(d *burnproof.CoinsProof) error
}

type pool struct {
	netParams *params.ChainParams
	log       logger.Logger
	ctx       context.Context

	chain chain.Blockchain
	host  hostnode.HostNode
	//lastActionManager actionmanager.LastActionManager

	votesLock sync.Mutex
	votes     map[chainhash.Hash]*primitives.MultiValidatorVote

	intidivualVotes     map[chainhash.Hash][]*primitives.MultiValidatorVote
	intidivualVotesLock sync.Mutex

	depositsLock sync.Mutex
	deposits     map[chainhash.Hash]*primitives.Deposit

	exitsLock sync.Mutex
	exits     map[chainhash.Hash]*primitives.Exit

	partialExitsLock sync.Mutex
	partialExits     map[chainhash.Hash]*primitives.PartialExit

	txsLock sync.Mutex
	txs     map[chainhash.Hash]*primitives.Tx

	multiSignatureTxsLock sync.Mutex
	multiSignatureTx      map[chainhash.Hash]*primitives.MultiSignatureTx

	voteSlashingLock sync.Mutex
	voteSlashings    []*primitives.VoteSlashing

	proposerSlashingLock sync.Mutex
	proposerSlashings    []*primitives.ProposerSlashing

	randaoSlashingLock sync.Mutex
	randaoSlashings    []*primitives.RANDAOSlashing

	governanceVoteLock sync.Mutex
	governanceVotes    map[chainhash.Hash]*primitives.GovernanceVote

	coinProofsLock sync.Mutex
	coinProofs     map[chainhash.Hash]*burnproof.CoinsProofSerializable
}

func (p *pool) AddVote(d *primitives.MultiValidatorVote) error {
	panic("implement me")
}

func (p *pool) AddDeposit(d *primitives.Deposit) error {
	panic("implement me")
}

func (p *pool) AddExit(d *primitives.Exit) error {
	panic("implement me")
}

func (p *pool) AddPartialExit(d *primitives.PartialExit) error {
	panic("implement me")
}

func (p *pool) AddTx(d *primitives.Tx) error {
	panic("implement me")
}

func (p *pool) AddMultiSignatureTx(d *primitives.MultiSignatureTx) error {
	panic("implement me")
}

func (p *pool) AddVoteSlashing(d *primitives.VoteSlashing) error {
	panic("implement me")
}

func (p *pool) AddProposerSlashing(d *primitives.ProposerSlashing) error {
	panic("implement me")
}

func (p *pool) AddRANDAOSlashing(d *primitives.RANDAOSlashing) error {
	panic("implement me")
}

func (p *pool) AddGovernanceVote(d *primitives.GovernanceVote) error {
	panic("implement me")
}

func (p *pool) AddCoinProof(d *burnproof.CoinsProof) error {
	panic("implement me")
}

var _ Pool = &pool{}

// Load fills up the pool from disk
func (p *pool) Load() {

}

// Start initializes the pool listeners
func (p *pool) Start() {
	p.Load()
}

// Stop closes listeners and save to disk
func (p *pool) Stop() {
	p.votesLock.Lock()
	p.intidivualVotesLock.Lock()
	p.depositsLock.Lock()
	p.exitsLock.Lock()
	p.partialExitsLock.Lock()
	p.txsLock.Lock()
	p.multiSignatureTxsLock.Lock()
	p.voteSlashingLock.Lock()
	p.proposerSlashingLock.Lock()
	p.randaoSlashingLock.Lock()
	p.governanceVoteLock.Lock()
	p.coinProofsLock.Lock()
	defer p.votesLock.Unlock()
	defer p.intidivualVotesLock.Unlock()
	defer p.depositsLock.Unlock()
	defer p.exitsLock.Unlock()
	defer p.partialExitsLock.Unlock()
	defer p.txsLock.Unlock()
	defer p.multiSignatureTxsLock.Unlock()
	defer p.voteSlashingLock.Unlock()
	defer p.proposerSlashingLock.Unlock()
	defer p.randaoSlashingLock.Unlock()
	defer p.governanceVoteLock.Unlock()
	defer p.coinProofsLock.Unlock()
}

func NewPool(ch chain.Blockchain, hostnode hostnode.HostNode) Pool {
	return &pool{
		netParams: config.GlobalParams.NetParams,
		log:       config.GlobalParams.Logger,
		ctx:       config.GlobalParams.Context,
		chain:     ch,
		host:      hostnode,
		//lastActionManager: manager,

		votes:             make(map[chainhash.Hash]*primitives.MultiValidatorVote),
		intidivualVotes:   make(map[chainhash.Hash][]*primitives.MultiValidatorVote),
		deposits:          make(map[chainhash.Hash]*primitives.Deposit),
		exits:             make(map[chainhash.Hash]*primitives.Exit),
		partialExits:      make(map[chainhash.Hash]*primitives.PartialExit),
		txs:               make(map[chainhash.Hash]*primitives.Tx),
		multiSignatureTx:  make(map[chainhash.Hash]*primitives.MultiSignatureTx),
		voteSlashings:     []*primitives.VoteSlashing{},
		proposerSlashings: []*primitives.ProposerSlashing{},
		randaoSlashings:   []*primitives.RANDAOSlashing{},
		governanceVotes:   make(map[chainhash.Hash]*primitives.GovernanceVote),
		coinProofs:        make(map[chainhash.Hash]*burnproof.CoinsProofSerializable),
	}
}
