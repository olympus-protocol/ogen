package mempool

import (
	"bytes"
	"context"
	"sync"

	"github.com/olympus-protocol/ogen/chain/index"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/logger"
)

// ActionMempool keeps track of actions to be added to the blockchain
// such as deposits, withdrawals, slashings, etc.
type ActionMempool struct {
	depositsLock sync.Mutex
	deposits     []primitives.Deposit

	exitsLock sync.Mutex
	exits     []primitives.Exit

	voteSlashingLock sync.Mutex
	voteSlashings    []primitives.VoteSlashing

	proposerSlashingLock sync.Mutex
	proposerSlashings    []primitives.ProposerSlashing

	randaoSlashingLock sync.Mutex
	randaoSlashings    []primitives.RANDAOSlashing

	governanceVoteLock sync.Mutex
	governanceVotes    []primitives.GovernanceVote

	params     *params.ChainParams
	ctx        context.Context
	log        *logger.Logger
	blockchain *chain.Blockchain
	hostNode   *peers.HostNode
}

func (am *ActionMempool) NotifyIllegalVotes(slashing primitives.VoteSlashing) {
	slot1 := slashing.Vote1.Data.Slot
	slot2 := slashing.Vote2.Data.Slot

	maxSlot := slot1
	if slot2 > slot1 {
		maxSlot = slot2
	}

	tipState, err := am.blockchain.State().TipStateAtSlot(maxSlot)
	if err != nil {
		am.log.Error(err)
		return
	}

	if _, err := tipState.IsVoteSlashingValid(&slashing, am.params); err != nil {
		am.log.Error(err)
		return
	}

	am.voteSlashingLock.Lock()
	defer am.voteSlashingLock.Unlock()

	sh := slashing.Hash()
	for _, d := range am.voteSlashings {
		dh := d.Hash()
		if dh.IsEqual(&sh) {
			return
		}
	}

	am.voteSlashings = append(am.voteSlashings, slashing)
}

func (am *ActionMempool) NewTip(_ *index.BlockRow, _ *primitives.Block, _ *primitives.State, _ []*primitives.EpochReceipt) {
}

func (am *ActionMempool) ProposerSlashingConditionViolated(slashing primitives.ProposerSlashing) {
	slot1 := slashing.BlockHeader1.Slot
	slot2 := slashing.BlockHeader2.Slot

	maxSlot := slot1
	if slot2 > slot1 {
		maxSlot = slot2
	}

	tipState, err := am.blockchain.State().TipStateAtSlot(maxSlot)
	if err != nil {
		am.log.Error(err)
		return
	}

	if _, err := tipState.IsProposerSlashingValid(&slashing); err != nil {
		am.log.Error(err)
		return
	}

	am.proposerSlashingLock.Lock()
	defer am.proposerSlashingLock.Unlock()

	sh := slashing.Hash()
	for _, d := range am.proposerSlashings {
		dh := d.Hash()
		if dh.IsEqual(&sh) {
			return
		}
	}

	am.proposerSlashings = append(am.proposerSlashings, slashing)
}

// NewActionMempool constructs a new action mempool.
func NewActionMempool(ctx context.Context, log *logger.Logger, p *params.ChainParams, blockchain *chain.Blockchain, hostnode *peers.HostNode) (*ActionMempool, error) {
	depositTopic, err := hostnode.Topic("deposits")
	if err != nil {
		return nil, err
	}

	depositTopicSub, err := depositTopic.Subscribe()
	if err != nil {
		return nil, err
	}

	exitTopic, err := hostnode.Topic("exits")
	if err != nil {
		return nil, err
	}

	exitTopicSub, err := exitTopic.Subscribe()
	if err != nil {
		return nil, err
	}

	governanceTopic, err := hostnode.Topic("governance")
	if err != nil {
		return nil, err
	}

	governanceTopicSub, err := governanceTopic.Subscribe()
	if err != nil {
		return nil, err
	}

	am := &ActionMempool{
		params:     p,
		ctx:        ctx,
		log:        log,
		blockchain: blockchain,
		hostNode:   hostnode,
	}

	blockchain.Notify(am)

	go am.handleDepositSub(depositTopicSub)
	go am.handleExitSub(exitTopicSub)
	go am.handleGovernanceSub(governanceTopicSub)

	return am, nil
}

func (am *ActionMempool) handleDepositSub(sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(am.ctx)
		if err != nil {
			am.log.Warnf("error getting next message in deposits topic: %s", err)
			return
		}

		txBuf := bytes.NewReader(msg.Data)
		tx := new(primitives.Deposit)

		if err := tx.Decode(txBuf); err != nil {
			// TODO: ban peer
			am.log.Warnf("peer sent invalid deposit: %s", err)
			continue
		}

		currentState := am.blockchain.State().TipState()

		err = am.AddDeposit(tx, currentState)
		if err != nil {
			am.log.Debugf("error adding transaction to mempool (might not be synced): %s", err)
		}
	}
}

// AddDeposit adds a deposit to the mempool.
func (am *ActionMempool) AddDeposit(deposit *primitives.Deposit, state *primitives.State) error {
	if err := state.IsDepositValid(deposit, am.params); err != nil {
		return err
	}

	am.depositsLock.Lock()
	defer am.depositsLock.Unlock()

	for _, d := range am.deposits {
		if bytes.Equal(d.Data.PublicKey.Marshal(), deposit.Data.PublicKey.Marshal()) {
			return nil
		}
	}

	am.deposits = append(am.deposits, *deposit)

	return nil
}

// GetDeposits gets deposits from the mempool. Mutates withState.
func (am *ActionMempool) GetDeposits(num int, withState *primitives.State) ([]primitives.Deposit, *primitives.State, error) {
	am.depositsLock.Lock()
	defer am.depositsLock.Unlock()
	deposits := make([]primitives.Deposit, 0, num)
	newMempool := make([]primitives.Deposit, 0, len(am.deposits))

	for _, d := range am.deposits {
		if err := withState.ApplyDeposit(&d, am.params); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool = append(newMempool, d)

		if len(deposits) < num {
			deposits = append(deposits, d)
		}
	}

	am.deposits = newMempool

	return deposits, withState, nil
}

// RemoveByBlock removes transactions that were in an accepted block.
func (am *ActionMempool) RemoveByBlock(b *primitives.Block, tipState *primitives.State) {
	am.depositsLock.Lock()
	newDeposits := make([]primitives.Deposit, 0, len(am.deposits))
outer:
	for _, d1 := range am.deposits {
		for _, d2 := range b.Deposits {
			if bytes.Equal(d1.Data.PublicKey.Marshal(), d2.Data.PublicKey.Marshal()) {
				continue outer
			}
		}

		if tipState.IsDepositValid(&d1, am.params) != nil {
			continue
		}

		newDeposits = append(newDeposits, d1)
	}
	am.deposits = newDeposits
	am.depositsLock.Unlock()

	am.exitsLock.Lock()
	newExits := make([]primitives.Exit, 0, len(am.exits))
outer1:
	for _, e1 := range am.exits {
		for _, e2 := range b.Exits {
			if bytes.Equal(e1.ValidatorPubkey.Marshal(), e2.ValidatorPubkey.Marshal()) {
				continue outer1
			}
		}

		if tipState.IsExitValid(&e1) != nil {
			continue
		}

		newExits = append(newExits, e1)
	}
	am.exits = newExits
	am.exitsLock.Unlock()

	am.proposerSlashingLock.Lock()
	newProposerSlashings := make([]primitives.ProposerSlashing, 0, len(am.proposerSlashings))
	for _, ps := range am.proposerSlashings {
		psHash := ps.Hash()
		if b.Header.Slot >= ps.BlockHeader2.Slot+am.params.EpochLength-1 {
			continue
		}

		if b.Header.Slot >= ps.BlockHeader1.Slot+am.params.EpochLength-1 {
			continue
		}

		for _, blockSlashing := range b.ProposerSlashings {
			blockSlashingHash := blockSlashing.Hash()

			if blockSlashingHash.IsEqual(&psHash) {
				continue
			}
		}

		if _, err := tipState.IsProposerSlashingValid(&ps); err != nil {
			continue
		}

		newProposerSlashings = append(newProposerSlashings, ps)
	}
	am.proposerSlashings = newProposerSlashings
	am.proposerSlashingLock.Unlock()

	am.voteSlashingLock.Lock()
	newVoteSlashings := make([]primitives.VoteSlashing, 0, len(am.voteSlashings))
	for _, vs := range am.voteSlashings {
		vsHash := vs.Hash()
		if b.Header.Slot >= vs.Vote1.Data.LastSlotValid(am.params) {
			continue
		}

		if b.Header.Slot >= vs.Vote2.Data.LastSlotValid(am.params) {
			continue
		}

		for _, voteSlashing := range b.VoteSlashings {
			voteSlashingHash := voteSlashing.Hash()

			if voteSlashingHash.IsEqual(&vsHash) {
				continue
			}
		}

		if _, err := tipState.IsVoteSlashingValid(&vs, am.params); err != nil {
			continue
		}

		newVoteSlashings = append(newVoteSlashings, vs)
	}
	am.voteSlashings = newVoteSlashings
	am.voteSlashingLock.Unlock()

	am.randaoSlashingLock.Lock()
	newRANDAOSlashings := make([]primitives.RANDAOSlashing, 0, len(am.randaoSlashings))
	for _, rs := range am.randaoSlashings {
		rsHash := rs.Hash()

		for _, blockSlashing := range b.VoteSlashings {
			blockSlashingHash := blockSlashing.Hash()

			if blockSlashingHash.IsEqual(&rsHash) {
				continue
			}
		}

		if _, err := tipState.IsRANDAOSlashingValid(&rs); err != nil {
			continue
		}

		newRANDAOSlashings = append(newRANDAOSlashings, rs)
	}
	am.randaoSlashings = newRANDAOSlashings
	am.randaoSlashingLock.Unlock()

	am.governanceVoteLock.Lock()
	newGovernanceVotes := make([]primitives.GovernanceVote, 0, len(am.governanceVotes))
	for _, gv := range am.governanceVotes {
		gvHash := gv.Hash()

		for _, blockSlashing := range b.VoteSlashings {
			blockSlashingHash := blockSlashing.Hash()

			if blockSlashingHash.IsEqual(&gvHash) {
				continue
			}
		}

		if err := tipState.IsGovernanceVoteValid(&gv, am.params); err != nil {
			continue
		}

		newGovernanceVotes = append(newGovernanceVotes, gv)
	}
	am.governanceVotes = newGovernanceVotes
	am.governanceVoteLock.Unlock()
}

func (am *ActionMempool) handleGovernanceSub(sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(am.ctx)
		if err != nil {
			am.log.Warnf("error getting next message in exits topic: %s", err)
			return
		}

		tx := new(primitives.GovernanceVote)

		if err := tx.Unmarshal(msg.Data); err != nil {
			// TODO: ban peer
			am.log.Warnf("peer sent invalid governance vote: %s", err)
			continue
		}

		currentState := am.blockchain.State().TipState()

		err = am.AddGovernanceVote(tx, currentState)
		if err != nil {
			am.log.Debugf("error adding transaction to mempool (might not be synced): %s", err)
		}
	}
}

// AddGovernanceVote adds a governance vote to the mempool.
func (am *ActionMempool) AddGovernanceVote(vote *primitives.GovernanceVote, state *primitives.State) error {
	if err := state.IsGovernanceVoteValid(vote, am.params); err != nil {
		return err
	}

	am.governanceVoteLock.Lock()
	defer am.governanceVoteLock.Unlock()

	voteHash := vote.Hash()

	for _, v := range am.governanceVotes {
		vh := v.Hash()
		if vh.IsEqual(&voteHash) {
			return nil
		}
	}

	am.governanceVotes = append(am.governanceVotes, *vote)

	return nil
}

func (am *ActionMempool) handleExitSub(sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(am.ctx)
		if err != nil {
			am.log.Warnf("error getting next message in exits topic: %s", err)
			return
		}

		tx := new(primitives.Exit)

		if err := tx.Unmarshal(msg.Data); err != nil {
			// TODO: ban peer
			am.log.Warnf("peer sent invalid exit: %s", err)
			continue
		}

		currentState := am.blockchain.State().TipState()

		err = am.AddExit(tx, currentState)
		if err != nil {
			am.log.Debugf("error adding transaction to mempool (might not be synced): %s", err)
		}
	}
}

// AddExit adds a exit to the mempool.
func (am *ActionMempool) AddExit(exit *primitives.Exit, state *primitives.State) error {
	if err := state.IsExitValid(exit); err != nil {
		return err
	}

	am.exitsLock.Lock()
	defer am.exitsLock.Unlock()

	for _, e := range am.exits {
		if bytes.Equal(e.ValidatorPubkey.Marshal(), e.ValidatorPubkey.Marshal()) {
			return nil
		}
	}

	am.exits = append(am.exits, *exit)

	return nil
}

// GetExits gets exits from the mempool. Mutates withState.
func (am *ActionMempool) GetExits(num int, state *primitives.State) ([]primitives.Exit, error) {
	am.exitsLock.Lock()
	defer am.exitsLock.Unlock()
	exits := make([]primitives.Exit, 0, num)
	newMempool := make([]primitives.Exit, 0, len(am.exits))

	for _, e := range am.exits {
		if err := state.ApplyExit(&e); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool = append(newMempool, e)

		if len(exits) < num {
			exits = append(exits, e)
		}
	}

	am.exits = newMempool

	return exits, nil
}

// GetProposerSlashings gets proposer slashings from the mempool. Mutates withState.
func (am *ActionMempool) GetProposerSlashings(num int, state *primitives.State) ([]primitives.ProposerSlashing, error) {
	am.proposerSlashingLock.Lock()
	defer am.proposerSlashingLock.Unlock()
	slashings := make([]primitives.ProposerSlashing, 0, num)
	newMempool := make([]primitives.ProposerSlashing, 0, len(am.proposerSlashings))

	for _, ps := range am.proposerSlashings {
		if err := state.ApplyProposerSlashing(&ps, am.params); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool = append(newMempool, ps)

		if len(slashings) < num {
			slashings = append(slashings, ps)
		}
	}

	am.proposerSlashings = newMempool

	return slashings, nil
}

// GetVoteSlashings gets vote slashings from the mempool. Mutates withState.
func (am *ActionMempool) GetVoteSlashings(num int, state *primitives.State) ([]primitives.VoteSlashing, error) {
	am.voteSlashingLock.Lock()
	defer am.voteSlashingLock.Unlock()
	slashings := make([]primitives.VoteSlashing, 0, num)
	newMempool := make([]primitives.VoteSlashing, 0, len(am.voteSlashings))

	for _, vs := range am.voteSlashings {
		if err := state.ApplyVoteSlashing(&vs, am.params); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool = append(newMempool, vs)

		if len(slashings) < num {
			slashings = append(slashings, vs)
		}
	}

	am.voteSlashings = newMempool

	return slashings, nil
}

// GetRANDAOSlashings gets RANDAO slashings from the mempool. Mutates withState.
func (am *ActionMempool) GetRANDAOSlashings(num int, state *primitives.State) ([]primitives.RANDAOSlashing, error) {
	am.randaoSlashingLock.Lock()
	defer am.randaoSlashingLock.Unlock()
	slashings := make([]primitives.RANDAOSlashing, 0, num)
	newMempool := make([]primitives.RANDAOSlashing, 0, len(am.randaoSlashings))

	for _, rs := range am.randaoSlashings {
		if err := state.ApplyRANDAOSlashing(&rs, am.params); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool = append(newMempool, rs)

		if len(slashings) < num {
			slashings = append(slashings, rs)
		}
	}

	am.randaoSlashings = newMempool

	return slashings, nil
}

// GetGovernanceVotes gets governance votes from the mempool. Mutates state.
func (am *ActionMempool) GetGovernanceVotes(num int, state *primitives.State) ([]primitives.GovernanceVote, error) {
	am.governanceVoteLock.Lock()
	defer am.governanceVoteLock.Unlock()
	votes := make([]primitives.GovernanceVote, 0, num)
	newMempool := make([]primitives.GovernanceVote, 0, len(am.governanceVotes))

	for _, gv := range am.governanceVotes {
		if err := state.ProcessGovernanceVote(&gv, am.params); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool = append(newMempool, gv)

		if len(votes) < num {
			votes = append(votes, gv)
		}
	}

	am.governanceVotes = newMempool

	return votes, nil
}

var _ chain.BlockchainNotifee = &ActionMempool{}
var _ VoteSlashingNotifee = &ActionMempool{}
