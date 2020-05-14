package mempool

import (
	"bytes"
	"context"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
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

	am := &ActionMempool{
		params:     p,
		ctx:        ctx,
		log:        log,
		blockchain: blockchain,
		hostNode:   hostnode,
	}

	go am.handleDepositSub(depositTopicSub)
	go am.handleExitSub(exitTopicSub)

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
			am.log.Warnf("error adding transaction to mempool: %s", err)
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

func (am *ActionMempool) handleExitSub(sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(am.ctx)
		if err != nil {
			am.log.Warnf("error getting next message in exits topic: %s", err)
			return
		}

		txBuf := bytes.NewReader(msg.Data)
		tx := new(primitives.Exit)

		if err := tx.Decode(txBuf); err != nil {
			// TODO: ban peer
			am.log.Warnf("peer sent invalid deposit: %s", err)
			continue
		}

		currentState := am.blockchain.State().TipState()

		err = am.AddExit(tx, currentState)
		if err != nil {
			am.log.Warnf("error adding transaction to mempool: %s", err)
		}
	}
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
