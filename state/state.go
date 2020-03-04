package state

import (
	"github.com/olympus-protocol/ogen/gov"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/users"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/workers"
)

type State struct {
	UtxoState       UtxoState
	GovernanceState GovernanceState
	UserState       UserState
	WorkerState     WorkerState
}

type GovernanceProposal struct {
	OutPoint p2p.OutPoint
	GovData  gov.GovObject
}

type GovernanceState struct {
	Proposals map[chainhash.Hash]GovernanceProposal
}

// Have checks if the governance state contains a specific proposal hash.
func (g *GovernanceState) Have(c chainhash.Hash) bool {
	_, ok := g.Proposals[c]
	return ok
}

// Get gets a governance proposal from the governance state.
func (g *GovernanceState) Get(c chainhash.Hash) GovernanceProposal {
	return g.Proposals[c]
}

type Utxo struct {
	OutPoint          p2p.OutPoint
	PrevInputsPubKeys [][48]byte
	Owner             string
	Amount            int64
}

type UtxoState struct {
	UTXOs map[chainhash.Hash]Utxo
}

// Have checks if a UTXO exists.
func (u *UtxoState) Have(c chainhash.Hash) bool {
	_, found := u.UTXOs[c]
	return found
}

// Get gets the UTXO from state.
func (u *UtxoState) Get(c chainhash.Hash) Utxo {
	return u.UTXOs[c]
}

func (s *State) TransitionBlock(block primitives.Block) State {
	return *s
}

type User struct {
	OutPoint p2p.OutPoint
	UserData users.User
}

type UserState struct {
	Users map[chainhash.Hash]User
}

// Have checks if a User exists.
func (u *UserState) Have(c chainhash.Hash) bool {
	_, found := u.Users[c]
	return found
}

// Get gets a User from state.
func (u *UserState) Get(c chainhash.Hash) User {
	return u.Users[c]
}

type Worker struct {
	OutPoint   p2p.OutPoint
	WorkerData workers.Worker
}

type WorkerState struct {
	Workers map[chainhash.Hash]Worker
}

// Have checks if a Worker exists.
func (u *WorkerState) Have(c chainhash.Hash) bool {
	_, found := u.Workers[c]
	return found
}

// Get gets a Worker from state.
func (u *WorkerState) Get(c chainhash.Hash) Worker {
	return u.Workers[c]
}
