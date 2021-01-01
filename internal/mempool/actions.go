package mempool

import (
	"bytes"
	"context"
	"errors"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"sync"

	"github.com/olympus-protocol/ogen/internal/chainindex"

	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// ActioMempool is the interface dor actionMempool
type ActionMempool interface {
	// Notifiers
	NotifyIllegalVotes(slashing *primitives.VoteSlashing)
	NewTip(_ *chainindex.BlockRow, _ *primitives.Block, _ state.State, _ []*primitives.EpochReceipt)
	ProposerSlashingConditionViolated(slashing *primitives.ProposerSlashing)

	// Remove on block
	RemoveByBlock(b *primitives.Block, tipState state.State)

	// Adders
	AddDeposit(deposit *primitives.Deposit) error
	AddGovernanceVote(vote *primitives.GovernanceVote, state state.State) error
	AddExit(exit *primitives.Exit) error
	AddProof(p *burnproof.CoinsProofSerializable) error
	AddPartialExit(p *primitives.PartialExit) error

	// Getters
	GetDeposits(num int, withState state.State) ([]*primitives.Deposit, state.State, error)
	GetProposerSlashings(num int, state state.State) ([]*primitives.ProposerSlashing, error)
	GetExits(num int, state state.State) ([]*primitives.Exit, error)
	GetVoteSlashings(num int, state state.State) ([]*primitives.VoteSlashing, error)
	GetRANDAOSlashings(num int, state state.State) ([]*primitives.RANDAOSlashing, error)
	GetGovernanceVotes(num int, state state.State) ([]*primitives.GovernanceVote, error)
	GetProofs(num int, state state.State) ([]*burnproof.CoinsProofSerializable, error)
	GetPartialExits(num int, state state.State) ([]*primitives.PartialExit, error)
}

var _ ActionMempool = &actionMempool{}

// ActionMempool keeps track of actions to be added to the blockchain
// such as deposits, withdrawals, slashings, etc.
type actionMempool struct {
	depositsLock sync.Mutex
	deposits     map[chainhash.Hash]*primitives.Deposit

	exitsLock sync.Mutex
	exits     map[chainhash.Hash]*primitives.Exit

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

	partialExitsLock sync.Mutex
	partialExits     map[chainhash.Hash]*primitives.PartialExit

	netParams *params.ChainParams
	ctx       context.Context
	log       logger.Logger
	chain     chain.Blockchain
	host      hostnode.HostNode
}

func (am *actionMempool) NotifyIllegalVotes(slashing *primitives.VoteSlashing) {
	slot1 := slashing.Vote1.Data.Slot
	slot2 := slashing.Vote2.Data.Slot

	maxSlot := slot1
	if slot2 > slot1 {
		maxSlot = slot2
	}

	tipState, err := am.chain.State().TipStateAtSlot(maxSlot)
	if err != nil {
		am.log.Error(err)
		return
	}

	if _, err := tipState.IsVoteSlashingValid(slashing); err != nil {
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

func (am *actionMempool) NewTip(_ *chainindex.BlockRow, _ *primitives.Block, _ state.State, _ []*primitives.EpochReceipt) {
}

func (am *actionMempool) ProposerSlashingConditionViolated(slashing *primitives.ProposerSlashing) {
	slot1 := slashing.BlockHeader1.Slot
	slot2 := slashing.BlockHeader2.Slot

	maxSlot := slot1
	if slot2 > slot1 {
		maxSlot = slot2
	}

	tipState, err := am.chain.State().TipStateAtSlot(maxSlot)
	if err != nil {
		am.log.Error(err)
		return
	}

	if _, err := tipState.IsProposerSlashingValid(slashing); err != nil {
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
func NewActionMempool(blockchain chain.Blockchain, hostnode hostnode.HostNode) (ActionMempool, error) {
	netParams := config.GlobalParams.NetParams
	ctx := config.GlobalParams.Context
	log := config.GlobalParams.Logger

	am := &actionMempool{
		netParams: netParams,
		ctx:       ctx,
		log:       log,
		chain:     blockchain,
		host:      hostnode,

		deposits:        make(map[chainhash.Hash]*primitives.Deposit),
		exits:           make(map[chainhash.Hash]*primitives.Exit),
		governanceVotes: make(map[chainhash.Hash]*primitives.GovernanceVote),
	}

	blockchain.Notify(am)

	if err := am.host.RegisterTopicHandler(p2p.MsgDepositCmd, am.handleDeposit); err != nil {
		return nil, err
	}

	if err := am.host.RegisterTopicHandler(p2p.MsgDepositsCmd, am.handleDeposits); err != nil {
		return nil, err
	}

	if err := am.host.RegisterTopicHandler(p2p.MsgExitCmd, am.handleExit); err != nil {
		return nil, err
	}

	if err := am.host.RegisterTopicHandler(p2p.MsgExitsCmd, am.handleExits); err != nil {
		return nil, err
	}

	if err := am.host.RegisterTopicHandler(p2p.MsgGovernanceCmd, am.handleGovernance); err != nil {
		return nil, err
	}

	if err := am.host.RegisterTopicHandler(p2p.MsgProofsCmd, am.handleProofs); err != nil {
		return nil, err
	}

	return am, nil
}

func (am *actionMempool) handleDeposit(id peer.ID, msg p2p.Message) error {
	if id == am.host.GetHost().ID() {
		return nil
	}
	data, ok := msg.(*p2p.MsgDeposit)
	if !ok {
		return errors.New("wrong message on deposit topic")
	}
	err := am.AddDeposit(data.Data)
	if err != nil {
		return err
	}

	return nil
}

func (am *actionMempool) handleDeposits(id peer.ID, msg p2p.Message) error {
	if id == am.host.GetHost().ID() {
		return nil
	}
	data, ok := msg.(*p2p.MsgDeposits)
	if !ok {
		return errors.New("wrong message on deposits topic")
	}

	for _, d := range data.Data {
		err := am.AddDeposit(d)
		if err != nil {
			return err
		}
	}

	return nil
}

func (am *actionMempool) handleExit(id peer.ID, msg p2p.Message) error {

	if id == am.host.GetHost().ID() {
		return nil
	}

	data, ok := msg.(*p2p.MsgExit)
	if !ok {
		return errors.New("wrong message on exit topic")
	}

	err := am.AddExit(data.Data)
	if err != nil {
		return err
	}

	return nil
}

func (am *actionMempool) handleExits(id peer.ID, msg p2p.Message) error {

	if id == am.host.GetHost().ID() {
		return nil
	}

	data, ok := msg.(*p2p.MsgExits)
	if !ok {
		return errors.New("wrong message on exits topic")
	}

	for _, d := range data.Data {

		err := am.AddExit(d)
		if err != nil {
			return err
		}

	}

	return nil
}

func (am *actionMempool) handleGovernance(id peer.ID, msg p2p.Message) error {

	if id == am.host.GetHost().ID() {
		return nil
	}

	data, ok := msg.(*p2p.MsgGovernance)
	if !ok {
		return errors.New("wrong message on governance topic")
	}

	s := am.chain.State().TipState()

	err := am.AddGovernanceVote(data.Data, s)
	if err != nil {
		return err
	}

	return nil
}

func (am *actionMempool) handleProofs(id peer.ID, msg p2p.Message) error {

	if id == am.host.GetHost().ID() {
		return nil
	}

	data, ok := msg.(*p2p.MsgProofs)
	if !ok {
		return errors.New("wrong message on proofs topic")
	}

	for _, p := range data.Proofs {
		err := am.AddProof(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (am *actionMempool) handlePartialExits(id peer.ID, msg p2p.Message) error {

	if id == am.host.GetHost().ID() {
		return nil
	}

	data, ok := msg.(*p2p.MsgPartialExits)
	if !ok {
		return errors.New("wrong message on proofs topic")
	}

	for _, p := range data.Data {
		err := am.AddPartialExit(p)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddDeposit adds a deposit to the mempool.
func (am *actionMempool) AddDeposit(deposit *primitives.Deposit) error {
	s := am.chain.State().TipState()

	if err := s.IsDepositValid(deposit); err != nil {
		return err
	}

	am.depositsLock.Lock()
	defer am.depositsLock.Unlock()

	for _, d := range am.deposits {
		if bytes.Equal(d.Data.PublicKey[:], deposit.Data.PublicKey[:]) {
			return nil
		}
	}
	_, ok := am.deposits[deposit.Hash()]
	if !ok {
		am.deposits[deposit.Hash()] = deposit
	}

	return nil
}

// AddPartialExit adds a PartialExits to the mempool.
func (am *actionMempool) AddPartialExit(p *primitives.PartialExit) error {
	s := am.chain.State().TipState()

	if err := s.IsPartialExitValid(p); err != nil {
		return err
	}

	am.partialExitsLock.Lock()
	defer am.partialExitsLock.Unlock()

	for _, pe := range am.partialExits {
		if bytes.Equal(pe.ValidatorPubkey[:], p.ValidatorPubkey[:]) && pe.Amount == p.Amount {
			return nil
		}
	}
	_, ok := am.partialExits[p.Hash()]
	if !ok {
		am.partialExits[p.Hash()] = p
	}

	return nil
}

// GetDeposits gets deposits from the mempool. Mutates withState.
func (am *actionMempool) GetDeposits(num int, withState state.State) ([]*primitives.Deposit, state.State, error) {
	am.depositsLock.Lock()
	defer am.depositsLock.Unlock()
	deposits := make([]*primitives.Deposit, 0, num)
	newMempool := make(map[chainhash.Hash]*primitives.Deposit)

	for k, d := range am.deposits {
		if err := withState.ApplyDeposit(d); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool[k] = d

		if len(deposits) < num {
			deposits = append(deposits, d)
		}
	}

	am.deposits = newMempool

	return deposits, withState, nil
}

// RemoveByBlock removes transactions that were in an accepted block.
func (am *actionMempool) RemoveByBlock(b *primitives.Block, tipState state.State) {
	am.depositsLock.Lock()
	newDeposits := make(map[chainhash.Hash]*primitives.Deposit)
outer:
	for k, d1 := range am.deposits {
		for _, d2 := range b.Deposits {
			if bytes.Equal(d1.Data.PublicKey[:], d2.Data.PublicKey[:]) {
				continue outer
			}
		}

		if tipState.IsDepositValid(d1) != nil {
			continue
		}

		newDeposits[k] = d1
	}
	am.deposits = newDeposits
	am.depositsLock.Unlock()

	am.exitsLock.Lock()
	newExits := make(map[chainhash.Hash]*primitives.Exit)
outer1:
	for k, e1 := range am.exits {
		for _, e2 := range b.Exits {
			if bytes.Equal(e1.ValidatorPubkey[:], e2.ValidatorPubkey[:]) {
				continue outer1
			}
		}

		if tipState.IsExitValid(e1) != nil {
			continue
		}

		newExits[k] = e1
	}
	am.exits = newExits
	am.exitsLock.Unlock()

	am.proposerSlashingLock.Lock()
	newProposerSlashings := make([]*primitives.ProposerSlashing, 0, len(am.proposerSlashings))
	for _, ps := range am.proposerSlashings {
		psHash := ps.Hash()
		if b.Header.Slot >= ps.BlockHeader2.Slot+am.netParams.EpochLength-1 {
			continue
		}

		if b.Header.Slot >= ps.BlockHeader1.Slot+am.netParams.EpochLength-1 {
			continue
		}

		for _, blockSlashing := range b.ProposerSlashings {
			blockSlashingHash := blockSlashing.Hash()

			if blockSlashingHash.IsEqual(&psHash) {
				continue
			}
		}

		if _, err := tipState.IsProposerSlashingValid(ps); err != nil {
			continue
		}

		newProposerSlashings = append(newProposerSlashings, ps)
	}
	am.proposerSlashings = newProposerSlashings
	am.proposerSlashingLock.Unlock()

	am.voteSlashingLock.Lock()
	newVoteSlashings := make([]*primitives.VoteSlashing, 0, len(am.voteSlashings))
	for _, vs := range am.voteSlashings {
		vsHash := vs.Hash()
		if b.Header.Slot >= vs.Vote1.Data.LastSlotValid(am.netParams) {
			continue
		}

		if b.Header.Slot >= vs.Vote2.Data.LastSlotValid(am.netParams) {
			continue
		}

		for _, voteSlashing := range b.VoteSlashings {
			voteSlashingHash := voteSlashing.Hash()

			if voteSlashingHash.IsEqual(&vsHash) {
				continue
			}
		}

		if _, err := tipState.IsVoteSlashingValid(vs); err != nil {
			continue
		}

		newVoteSlashings = append(newVoteSlashings, vs)
	}
	am.voteSlashings = newVoteSlashings
	am.voteSlashingLock.Unlock()

	am.randaoSlashingLock.Lock()
	newRANDAOSlashings := make([]*primitives.RANDAOSlashing, 0, len(am.randaoSlashings))
	for _, rs := range am.randaoSlashings {
		rsHash := rs.Hash()

		for _, blockSlashing := range b.VoteSlashings {
			blockSlashingHash := blockSlashing.Hash()

			if blockSlashingHash.IsEqual(&rsHash) {
				continue
			}
		}

		if _, err := tipState.IsRANDAOSlashingValid(rs); err != nil {
			continue
		}

		newRANDAOSlashings = append(newRANDAOSlashings, rs)
	}
	am.randaoSlashings = newRANDAOSlashings
	am.randaoSlashingLock.Unlock()

	am.governanceVoteLock.Lock()
	newGovernanceVotes := make(map[chainhash.Hash]*primitives.GovernanceVote)
	for k, gv := range am.governanceVotes {
		gvHash := gv.Hash()

		for _, blockSlashing := range b.VoteSlashings {
			blockSlashingHash := blockSlashing.Hash()

			if blockSlashingHash.IsEqual(&gvHash) {
				continue
			}
		}

		if err := tipState.IsGovernanceVoteValid(gv); err != nil {
			continue
		}

		newGovernanceVotes[k] = gv
	}
	am.governanceVotes = newGovernanceVotes
	am.governanceVoteLock.Unlock()

	am.coinProofsLock.Lock()
	newProofsPool := make(map[chainhash.Hash]*burnproof.CoinsProofSerializable)

	for k, proof := range am.coinProofs {
		proofHash := proof.Hash()

		for _, proof := range b.CoinProofs {
			blockProofHash := proof.Hash()

			if blockProofHash.IsEqual(&proofHash) {
				continue
			}
		}

		if err := tipState.IsCoinProofValid(proof); err != nil {
			continue
		}

		newProofsPool[k] = proof
	}
	am.coinProofs = newProofsPool
	am.coinProofsLock.Unlock()

	am.partialExitsLock.Lock()
	newPartialExitsPool := make(map[chainhash.Hash]*primitives.PartialExit)

	for k, exit := range am.partialExits {
		hash := exit.Hash()

		for _, e := range b.PartialExit {
			blockProofHash := e.Hash()

			if blockProofHash.IsEqual(&hash) {
				continue
			}
		}

		if err := tipState.IsPartialExitValid(exit); err != nil {
			continue
		}

		newPartialExitsPool[k] = exit
	}
	am.partialExits = newPartialExitsPool
	am.partialExitsLock.Unlock()

}

// AddGovernanceVote adds a governance vote to the mempool.
func (am *actionMempool) AddGovernanceVote(vote *primitives.GovernanceVote, state state.State) error {
	if err := state.IsGovernanceVoteValid(vote); err != nil {
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
	_, ok := am.governanceVotes[vote.Hash()]
	if !ok {
		am.governanceVotes[vote.Hash()] = vote
	}

	return nil
}

// AddExit adds a exit to the mempool.
func (am *actionMempool) AddExit(exit *primitives.Exit) error {
	s := am.chain.State().TipState()

	if err := s.IsExitValid(exit); err != nil {
		return err
	}

	am.exitsLock.Lock()
	defer am.exitsLock.Unlock()

	for _, e := range am.exits {
		if bytes.Equal(e.ValidatorPubkey[:], exit.ValidatorPubkey[:]) {
			return nil
		}
	}

	_, ok := am.exits[exit.Hash()]
	if !ok {
		am.exits[exit.Hash()] = exit
	}

	return nil
}

func (am *actionMempool) AddProof(p *burnproof.CoinsProofSerializable) error {
	s := am.chain.State().TipState()

	if err := s.IsCoinProofValid(p); err != nil {
		return err
	}

	am.coinProofsLock.Lock()
	defer am.coinProofsLock.Unlock()

	_, ok := am.coinProofs[p.Hash()]
	if !ok {
		am.coinProofs[p.Hash()] = p
	}

	return nil
}

// GetExits gets exits from the mempool. Mutates withState.
func (am *actionMempool) GetExits(num int, state state.State) ([]*primitives.Exit, error) {
	am.exitsLock.Lock()
	defer am.exitsLock.Unlock()
	exits := make([]*primitives.Exit, 0, num)
	newMempool := make(map[chainhash.Hash]*primitives.Exit)

	for k, e := range am.exits {
		if err := state.ApplyExit(e); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool[k] = e

		if len(exits) < num {
			exits = append(exits, e)
		}
	}

	am.exits = newMempool

	return exits, nil
}

// GetProposerSlashings gets proposer slashings from the mempool. Mutates withState.
func (am *actionMempool) GetProposerSlashings(num int, state state.State) ([]*primitives.ProposerSlashing, error) {
	am.proposerSlashingLock.Lock()
	defer am.proposerSlashingLock.Unlock()
	slashings := make([]*primitives.ProposerSlashing, 0, num)
	newMempool := make([]*primitives.ProposerSlashing, 0, len(am.proposerSlashings))

	for _, ps := range am.proposerSlashings {
		if err := state.ApplyProposerSlashing(ps); err != nil {
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
func (am *actionMempool) GetVoteSlashings(num int, state state.State) ([]*primitives.VoteSlashing, error) {
	am.voteSlashingLock.Lock()
	defer am.voteSlashingLock.Unlock()
	slashings := make([]*primitives.VoteSlashing, 0, num)
	newMempool := make([]*primitives.VoteSlashing, 0, len(am.voteSlashings))

	for _, vs := range am.voteSlashings {
		if err := state.ApplyVoteSlashing(vs); err != nil {
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
func (am *actionMempool) GetRANDAOSlashings(num int, state state.State) ([]*primitives.RANDAOSlashing, error) {
	am.randaoSlashingLock.Lock()
	defer am.randaoSlashingLock.Unlock()
	slashings := make([]*primitives.RANDAOSlashing, 0, num)
	newMempool := make([]*primitives.RANDAOSlashing, 0, len(am.randaoSlashings))

	for _, rs := range am.randaoSlashings {
		if err := state.ApplyRANDAOSlashing(rs); err != nil {
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
func (am *actionMempool) GetGovernanceVotes(num int, state state.State) ([]*primitives.GovernanceVote, error) {
	am.governanceVoteLock.Lock()
	defer am.governanceVoteLock.Unlock()
	votes := make([]*primitives.GovernanceVote, 0, num)
	newMempool := make(map[chainhash.Hash]*primitives.GovernanceVote)

	for k, gv := range am.governanceVotes {
		if err := state.ProcessGovernanceVote(gv); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool[k] = gv

		if len(votes) < num {
			votes = append(votes, gv)
		}
	}

	am.governanceVotes = newMempool

	return votes, nil
}

// GetProofs gets redeem proofs to be included on the block.
func (am *actionMempool) GetProofs(num int, state state.State) ([]*burnproof.CoinsProofSerializable, error) {
	am.coinProofsLock.Lock()
	defer am.coinProofsLock.Unlock()
	proofs := make([]*burnproof.CoinsProofSerializable, 0, num)
	newMempool := make(map[chainhash.Hash]*burnproof.CoinsProofSerializable)

	for k, p := range am.coinProofs {
		if err := state.ApplyCoinProof(p); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool[k] = p

		if len(proofs) < num {
			proofs = append(proofs, p)
		}
	}

	am.coinProofs = newMempool

	return proofs, nil
}

// GetPartialExits gets partial exits to be included on the block.
func (am *actionMempool) GetPartialExits(num int, state state.State) ([]*primitives.PartialExit, error) {
	am.partialExitsLock.Lock()
	defer am.partialExitsLock.Unlock()
	pexits := make([]*primitives.PartialExit, 0, num)

	newMempool := make(map[chainhash.Hash]*primitives.PartialExit)

	for k, p := range am.partialExits {
		if err := state.ApplyPartialExit(p); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool[k] = p

		if len(pexits) < num {
			pexits = append(pexits, p)
		}
	}

	am.partialExits = newMempool

	return pexits, nil
}

var _ chain.BlockchainNotifee = &actionMempool{}
var _ VoteSlashingNotifee = &actionMempool{}
