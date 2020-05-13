package mempool

import (
	"context"
	"sync"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/primitives"
)

// ActionMempool keeps track of actions to be added to the blockchain
// such as deposits, withdrawals, slashings, etc.
type ActionMempool struct {
	depositsLock sync.Mutex
	deposits     []primitives.Deposit

	exitsLock sync.Mutex
	exits     []primitives.Exit

	params     *params.ChainParams
	ctx        context.Context
	log        *logger.Logger
	blockchain *chain.Blockchain
	hostNode   *peers.HostNode
}

// NewActionMempool constructs a new action mempool.
func NewActionMempool(ctx context.Context, log *logger.Logger, p *params.ChainParams, blockchain *chain.Blockchain, hostnode *peers.HostNode) (*ActionMempool, error) {
	return &ActionMempool{
		params:     p,
		ctx:        ctx,
		log:        log,
		blockchain: blockchain,
		hostNode:   hostnode,
	}, nil
}

// AddDeposit adds a deposit to the mempool.
func (am *ActionMempool) AddDeposit(deposit *primitives.Deposit, state *primitives.State) error {
	if err := state.IsDepositValid(deposit, am.params); err != nil {
		return err
	}

	am.depositsLock.Lock()
	defer am.depositsLock.Unlock()

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

// AddExit adds a deposit to the mempool.
func (am *ActionMempool) AddExit(exit *primitives.Exit, state *primitives.State) error {
	if err := state.IsExitValid(exit); err != nil {
		return err
	}

	am.exitsLock.Lock()
	defer am.exitsLock.Unlock()

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
