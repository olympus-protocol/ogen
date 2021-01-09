package mempool

import (
	"context"
	"errors"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"sync"
)

var (
	ErrorAccountNotOnMempool = errors.New("account not on pool")
)

type Pool interface {
	Load()
	Start()
	Stop()

	AddVote(d *primitives.MultiValidatorVote, s state.State) error
	AddDeposit(d *primitives.Deposit) error
	AddExit(d *primitives.Exit) error
	AddPartialExit(d *primitives.PartialExit) error
	AddTx(d *primitives.Tx) error
	AddMultiSignatureTx(d *primitives.MultiSignatureTx) error
	AddVoteSlashing(d *primitives.VoteSlashing) error
	AddProposerSlashing(d *primitives.ProposerSlashing) error
	AddRANDAOSlashing(d *primitives.RANDAOSlashing) error
	AddGovernanceVote(d *primitives.GovernanceVote) error
	AddCoinProof(d *burnproof.CoinsProofSerializable) error

	GetAccountNonce(account [20]byte) (uint64, error)
	GetVotes(slotToPropose uint64, s state.State, index uint64) []*primitives.MultiValidatorVote
	GetDeposits(s state.State) ([]*primitives.Deposit, state.State)
	GetExits(s state.State) []*primitives.Exit
	GetPartialExits(s state.State) []*primitives.PartialExit
	GetCoinProofs(s state.State) []*burnproof.CoinsProofSerializable
	GetTxs(s state.State, feeReceiver [20]byte) ([]*primitives.Tx, state.State)
	GetVoteSlashings(s state.State) []*primitives.VoteSlashing
	GetProposerSlashings(s state.State) []*primitives.ProposerSlashing
	GetRANDAOSlashings(s state.State) []*primitives.RANDAOSlashing
	GetGovernanceVotes(s state.State) []*primitives.GovernanceVote
	GetMultiSignatureTxs(s state.State, feeReceiver [20]byte) ([]*primitives.MultiSignatureTx, state.State)

	RemoveByBlock(b *primitives.Block)
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

func (p *pool) AddVote(d *primitives.MultiValidatorVote, s state.State) error {
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

func (p *pool) AddCoinProof(d *burnproof.CoinsProofSerializable) error {
	panic("implement me")
}

func (p *pool) GetAccountNonce(account [20]byte) (uint64, error) {
	panic("implement me")
}

func (p *pool) GetVotes(slotToPropose uint64, s state.State, index uint64) []*primitives.MultiValidatorVote {
	panic("implement me")
}

func (p *pool) GetDeposits(s state.State) ([]*primitives.Deposit, state.State) {
	panic("implement me")
}

func (p *pool) GetExits(s state.State) []*primitives.Exit {
	panic("implement me")
}

func (p *pool) GetPartialExits(s state.State) []*primitives.PartialExit {
	panic("implement me")
}

func (p *pool) GetCoinProofs(s state.State) []*burnproof.CoinsProofSerializable {
	panic("implement me")
}

func (p *pool) GetTxs(s state.State, feeReceiver [20]byte) ([]*primitives.Tx, state.State) {
	panic("implement me")
}

func (p *pool) GetVoteSlashings(s state.State) []*primitives.VoteSlashing {
	panic("implement me")
}

func (p *pool) GetProposerSlashings(s state.State) []*primitives.ProposerSlashing {
	panic("implement me")
}

func (p *pool) GetRANDAOSlashings(s state.State) []*primitives.RANDAOSlashing {
	panic("implement me")
}

func (p *pool) GetGovernanceVotes(s state.State) []*primitives.GovernanceVote {
	panic("implement me")
}

func (p *pool) GetMultiSignatureTxs(s state.State, feeReceiver [20]byte) ([]*primitives.MultiSignatureTx, state.State) {
	panic("implement me")
}

func (p *pool) RemoveByBlock(b *primitives.Block) {
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
