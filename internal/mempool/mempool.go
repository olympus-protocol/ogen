package mempool

import (
	"bytes"
	"context"
	"errors"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"sync"
)

var (
	ErrorAccountNotOnMempool = errors.New("account not on pool")
)

type Pool interface {
	Load()
	Start() error
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
	GetExits(s state.State) ([]*primitives.Exit, state.State)
	GetPartialExits(s state.State) ([]*primitives.PartialExit, state.State)
	GetCoinProofs(s state.State) ([]*burnproof.CoinsProofSerializable, state.State)
	GetTxs(s state.State, feeReceiver [20]byte) ([]*primitives.Tx, state.State)
	GetVoteSlashings(s state.State) ([]*primitives.VoteSlashing, state.State)
	GetProposerSlashings(s state.State) ([]*primitives.ProposerSlashing, state.State)
	GetRANDAOSlashings(s state.State) ([]*primitives.RANDAOSlashing, state.State)
	GetGovernanceVotes(s state.State) ([]*primitives.GovernanceVote, state.State)
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

	latestNonceLock sync.Mutex
	latestNonce     map[[20]byte]uint64

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
	if err := s.IsVoteValid(d); err != nil {
		return err
	}

	p.votesLock.Lock()
	p.intidivualVotesLock.Lock()
	defer p.votesLock.Unlock()
	defer p.intidivualVotesLock.Unlock()

	voteData := d.Data
	voteHash := voteData.Hash()

	firstSlotAllowedToInclude := d.Data.Slot + p.netParams.MinAttestationInclusionDelay

	currentState, err := p.chain.State().TipStateAtSlot(firstSlotAllowedToInclude)
	if err != nil {
		p.log.Error(err)
		return err
	}

	// Register voting action for validators included on the vote
	//for i, c := range committee {
	//if vote.ParticipationBitfield.Get(uint(i)) {
	//m.lastActionManager.RegisterAction(currentState.GetValidatorRegistry()[c].PubKey, vote.Data.Nonce)
	//}
	//}

	// Slashing check
	// This check iterates over all the votes on the pool.
	// Checks if the new vote data matches any pool vote data hash.
	// If that check fails, we should check for validators submitting twice different votes.
	for h, v := range p.votes {

		// If the vote data hash matches, it means is voting for same block.
		if voteHash.IsEqual(&h) {
			continue
		}

		if currentState.GetSlot() >= v.Data.LastSlotValid(p.netParams) {
			delete(p.votes, voteHash)
			continue
		}

		var votingValidators = make(map[uint64]struct{})
		var intersect []uint64

		vote1Committee, err := currentState.GetVoteCommittee(v.Data.Slot)
		if err != nil {
			p.log.Error(err)
			return err
		}

		vote2Committee, err := currentState.GetVoteCommittee(d.Data.Slot)
		if err != nil {
			p.log.Error(err)
			return err
		}

		for i, idx := range vote1Committee {
			if !v.ParticipationBitfield.Get(uint(i)) {
				continue
			}
			votingValidators[idx] = struct{}{}
		}

		for i, idx := range vote2Committee {
			if !d.ParticipationBitfield.Get(uint(i)) {
				continue
			}
			_, exist := votingValidators[idx]
			if exist {
				intersect = append(intersect, idx)
			}
		}

		// Check if the new vote with different hash overlaps a previous marked validator vote.
		if len(intersect) != 0 {
			// Check if the vote matches the double vote and surround vote conditions
			// If there is an intersection check if is a double vote
			if v.Data.IsSurroundVote(voteData) || v.Data.IsDoubleVote(voteData) {
				// If is a double or surround vote announce it and slash.
				if v.Data.IsSurroundVote(voteData) {
					p.log.Warnf("found surround vote for multivalidator, reporting...")
				}
				if v.Data.IsDoubleVote(voteData) {
					p.log.Warnf("found double vote for multivalidator, reporting...")
				}
				vs := &primitives.VoteSlashing{
					Vote1: d,
					Vote2: v,
				}
				err = p.AddVoteSlashing(vs)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}

	// Check if vote is already on pool.
	// If a vote with same vote data is found we should check the signatures.
	// If the signatures are the same it means is a duplicated vote for network (probably a relayed vote).
	// If the signatures don't match, we should aggregate both signatures and merge the bitlists.
	// IMPORTANT: 	We should never allow a vote that conflicts a previous vote to be added to the pool.
	// 				That should be checked against all votes on pool comparing bitlists.
	v, ok := p.votes[voteHash]
	if ok {
		p.log.Debugf("received vote with same vote data aggregating %d votes...", len(d.ParticipationBitfield.BitIndices()))
		if !bytes.Equal(v.Sig[:], d.Sig[:]) {

			// Check if votes overlaps voters
			voteCommittee, err := currentState.GetVoteCommittee(v.Data.Slot)
			if err != nil {
				p.log.Error(err)
				return err
			}

			var sigs []uint64

			for i, idx := range voteCommittee {
				if v.ParticipationBitfield.Get(uint(i)) && d.ParticipationBitfield.Get(uint(i)) {
					sigs = append(sigs, idx)
				}
			}

			if len(sigs) != 0 {
				vs := &primitives.VoteSlashing{
					Vote1: d,
					Vote2: v,
				}
				err = p.AddVoteSlashing(vs)
				if err != nil {
					return err
				}
				return nil
			}

			newBitfield, err := v.ParticipationBitfield.Merge(d.ParticipationBitfield)
			if err != nil {
				p.log.Error(err)
				return err
			}

			sig1, err := bls.SignatureFromBytes(v.Sig[:])
			if err != nil {
				p.log.Error(err)
				return err
			}

			sig2, err := bls.SignatureFromBytes(d.Sig[:])
			if err != nil {
				p.log.Error(err)
				return err
			}

			newVoteSig := bls.AggregateSignatures([]common.Signature{sig1, sig2})

			var voteSig [96]byte
			copy(voteSig[:], newVoteSig.Marshal())

			newVote := &primitives.MultiValidatorVote{
				Data:                  v.Data,
				ParticipationBitfield: newBitfield,
				Sig:                   voteSig,
			}

			p.votes[voteHash] = newVote
			p.intidivualVotes[voteHash] = append(p.intidivualVotes[voteHash], d)
		}
	} else {
		p.log.Debugf("adding vote to the mempool with %d votes", len(d.ParticipationBitfield.BitIndices()))
		p.votes[voteHash] = d
		p.intidivualVotes[voteHash] = []*primitives.MultiValidatorVote{d}
	}

	return nil
}

func (p *pool) AddDeposit(d *primitives.Deposit) error {
	s := p.chain.State().TipState()

	if err := s.IsDepositValid(d); err != nil {
		return err
	}

	p.depositsLock.Lock()
	defer p.depositsLock.Unlock()

	for _, d := range p.deposits {
		if bytes.Equal(d.Data.PublicKey[:], d.Data.PublicKey[:]) {
			return nil
		}
	}
	_, ok := p.deposits[d.Hash()]
	if !ok {
		p.deposits[d.Hash()] = d
	}

	return nil
}

func (p *pool) AddExit(d *primitives.Exit) error {
	s := p.chain.State().TipState()

	if err := s.IsExitValid(d); err != nil {
		return err
	}

	p.exitsLock.Lock()
	defer p.exitsLock.Unlock()

	for _, e := range p.exits {
		if bytes.Equal(e.ValidatorPubkey[:], d.ValidatorPubkey[:]) {
			return nil
		}
	}

	_, ok := p.exits[d.Hash()]
	if !ok {
		p.exits[d.Hash()] = d
	}

	return nil
}

func (p *pool) AddPartialExit(d *primitives.PartialExit) error {
	s := p.chain.State().TipState()

	if err := s.IsPartialExitValid(d); err != nil {
		return err
	}

	p.partialExitsLock.Lock()
	defer p.partialExitsLock.Unlock()

	for _, pe := range p.partialExits {
		if bytes.Equal(pe.ValidatorPubkey[:], d.ValidatorPubkey[:]) {
			return nil
		}
	}
	_, ok := p.partialExits[d.Hash()]
	if !ok {
		p.partialExits[d.Hash()] = d
	}

	return nil
}

func (p *pool) AddTx(d *primitives.Tx) error {
	panic("implement me")
}

func (p *pool) AddMultiSignatureTx(d *primitives.MultiSignatureTx) error {
	panic("implement me")
}

func (p *pool) AddVoteSlashing(d *primitives.VoteSlashing) error {
	slot1 := d.Vote1.Data.Slot
	slot2 := d.Vote2.Data.Slot

	maxSlot := slot1
	if slot2 > slot1 {
		maxSlot = slot2
	}

	tipState, err := p.chain.State().TipStateAtSlot(maxSlot)
	if err != nil {
		p.log.Error(err)
		return err
	}

	if _, err := tipState.IsVoteSlashingValid(d); err != nil {
		p.log.Error(err)
		return err
	}

	p.voteSlashingLock.Lock()
	defer p.voteSlashingLock.Unlock()

	sh := d.Hash()
	for _, d := range p.voteSlashings {
		dh := d.Hash()
		if dh.IsEqual(&sh) {
			return nil
		}
	}

	p.voteSlashings = append(p.voteSlashings, d)

	return nil
}

func (p *pool) AddProposerSlashing(d *primitives.ProposerSlashing) error {
	slot1 := d.BlockHeader1.Slot
	slot2 := d.BlockHeader2.Slot

	maxSlot := slot1
	if slot2 > slot1 {
		maxSlot = slot2
	}

	tipState, err := p.chain.State().TipStateAtSlot(maxSlot)
	if err != nil {
		p.log.Error(err)
		return err
	}

	if _, err := tipState.IsProposerSlashingValid(d); err != nil {
		p.log.Error(err)
		return nil
	}

	p.proposerSlashingLock.Lock()
	defer p.proposerSlashingLock.Unlock()

	sh := d.Hash()
	for _, d := range p.proposerSlashings {
		dh := d.Hash()
		if dh.IsEqual(&sh) {
			return nil
		}
	}

	p.proposerSlashings = append(p.proposerSlashings, d)

	return nil
}

func (p *pool) AddRANDAOSlashing(_ *primitives.RANDAOSlashing) error {
	//panic("implement me")
	return nil
}

func (p *pool) AddGovernanceVote(d *primitives.GovernanceVote) error {
	s := p.chain.State().TipState()

	if err := s.IsGovernanceVoteValid(d); err != nil {
		return err
	}

	p.governanceVoteLock.Lock()
	defer p.governanceVoteLock.Unlock()

	voteHash := d.Hash()

	for _, v := range p.governanceVotes {
		vh := v.Hash()
		if vh.IsEqual(&voteHash) {
			return nil
		}
	}
	_, ok := p.governanceVotes[d.Hash()]
	if !ok {
		p.governanceVotes[d.Hash()] = d
	}

	return nil
}

func (p *pool) AddCoinProof(d *burnproof.CoinsProofSerializable) error {
	s := p.chain.State().TipState()

	if err := s.IsCoinProofValid(d); err != nil {
		return err
	}

	p.coinProofsLock.Lock()
	defer p.coinProofsLock.Unlock()

	_, ok := p.coinProofs[d.Hash()]
	if !ok {
		p.coinProofs[d.Hash()] = d
	}

	return nil
}

func (p *pool) GetAccountNonce(pkh [20]byte) (uint64, error) {
	p.latestNonceLock.Lock()
	defer p.latestNonceLock.Unlock()

	nonce, ok := p.latestNonce[pkh]
	if !ok {
		return 0, ErrorAccountNotOnMempool
	}

	return nonce, nil
}

func (p *pool) GetVotes(slotToPropose uint64, s state.State, index uint64) []*primitives.MultiValidatorVote {
	p.votesLock.Lock()
	defer p.votesLock.Unlock()

	votes := make([]*primitives.MultiValidatorVote, 0)

	for _, vote := range p.votes {

		if slotToPropose >= vote.Data.FirstSlotValid(p.netParams) && slotToPropose <= vote.Data.LastSlotValid(p.netParams) {
			err := s.ProcessVote(vote, index)
			if err != nil {
				p.log.Error(err)
				voteHash := vote.Data.Hash()
				delete(p.votes, voteHash)
				continue
			}
			if uint64(len(votes)) < primitives.MaxVotesPerBlock {
				votes = append(votes, vote)
			}
		}

	}

	return votes
}

func (p *pool) GetDeposits(s state.State) ([]*primitives.Deposit, state.State) {
	p.depositsLock.Lock()
	defer p.depositsLock.Unlock()
	deposits := make([]*primitives.Deposit, 0, primitives.MaxDepositsPerBlock)
	newMempool := make(map[chainhash.Hash]*primitives.Deposit)

	for k, d := range p.deposits {
		if err := s.ApplyDeposit(d); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool[k] = d

		if len(deposits) < primitives.MaxDepositsPerBlock {
			deposits = append(deposits, d)
		}
	}

	p.deposits = newMempool

	return deposits, s
}

func (p *pool) GetExits(s state.State) ([]*primitives.Exit, state.State) {
	p.exitsLock.Lock()
	defer p.exitsLock.Unlock()
	exits := make([]*primitives.Exit, 0, primitives.MaxExitsPerBlock)
	newMempool := make(map[chainhash.Hash]*primitives.Exit)

	for k, e := range p.exits {
		if err := s.ApplyExit(e); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool[k] = e

		if len(exits) < primitives.MaxExitsPerBlock {
			exits = append(exits, e)
		}
	}

	p.exits = newMempool

	return exits, s
}

func (p *pool) GetPartialExits(s state.State) ([]*primitives.PartialExit, state.State) {
	p.partialExitsLock.Lock()
	defer p.partialExitsLock.Unlock()
	pexits := make([]*primitives.PartialExit, 0, primitives.MaxPartialExitsPerBlock)

	newMempool := make(map[chainhash.Hash]*primitives.PartialExit)

	for k, p := range p.partialExits {
		if err := s.ApplyPartialExit(p); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool[k] = p

		if len(pexits) < primitives.MaxPartialExitsPerBlock {
			pexits = append(pexits, p)
		}
	}

	p.partialExits = newMempool

	return pexits, s
}

func (p *pool) GetCoinProofs(s state.State) ([]*burnproof.CoinsProofSerializable, state.State) {
	p.coinProofsLock.Lock()
	defer p.coinProofsLock.Unlock()
	proofs := make([]*burnproof.CoinsProofSerializable, 0, primitives.MaxCoinProofsPerBlock)
	newMempool := make(map[chainhash.Hash]*burnproof.CoinsProofSerializable)

	for k, p := range p.coinProofs {
		if err := s.ApplyCoinProof(p); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool[k] = p

		if len(proofs) < primitives.MaxCoinProofsPerBlock {
			proofs = append(proofs, p)
		}
	}

	p.coinProofs = newMempool

	return proofs, s
}

func (p *pool) GetTxs(s state.State, feeReceiver [20]byte) ([]*primitives.Tx, state.State) {
	panic("implement me")
}

func (p *pool) GetVoteSlashings(s state.State) ([]*primitives.VoteSlashing, state.State) {
	p.voteSlashingLock.Lock()
	defer p.voteSlashingLock.Unlock()
	slashings := make([]*primitives.VoteSlashing, 0, primitives.MaxVoteSlashingsPerBlock)
	newMempool := make([]*primitives.VoteSlashing, 0, len(p.voteSlashings))

	for _, vs := range p.voteSlashings {
		if err := s.ApplyVoteSlashing(vs); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool = append(newMempool, vs)

		if len(slashings) < primitives.MaxVoteSlashingsPerBlock {
			slashings = append(slashings, vs)
		}
	}

	p.voteSlashings = newMempool

	return slashings, s
}

func (p *pool) GetProposerSlashings(s state.State) ([]*primitives.ProposerSlashing, state.State) {
	p.proposerSlashingLock.Lock()
	defer p.proposerSlashingLock.Unlock()
	slashings := make([]*primitives.ProposerSlashing, 0, primitives.MaxProposerSlashingsPerBlock)
	newMempool := make([]*primitives.ProposerSlashing, 0, len(p.proposerSlashings))

	for _, ps := range p.proposerSlashings {
		if err := s.ApplyProposerSlashing(ps); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool = append(newMempool, ps)

		if len(slashings) < primitives.MaxProposerSlashingsPerBlock {
			slashings = append(slashings, ps)
		}
	}

	p.proposerSlashings = newMempool

	return slashings, nil
}

func (p *pool) GetRANDAOSlashings(s state.State) ([]*primitives.RANDAOSlashing, state.State) {
	p.randaoSlashingLock.Lock()
	defer p.randaoSlashingLock.Unlock()
	slashings := make([]*primitives.RANDAOSlashing, 0, primitives.MaxRANDAOSlashingsPerBlock)
	newMempool := make([]*primitives.RANDAOSlashing, 0, len(p.randaoSlashings))

	for _, rs := range p.randaoSlashings {
		if err := s.ApplyRANDAOSlashing(rs); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool = append(newMempool, rs)

		if len(slashings) < primitives.MaxRANDAOSlashingsPerBlock {
			slashings = append(slashings, rs)
		}
	}

	p.randaoSlashings = newMempool

	return slashings, s
}

func (p *pool) GetGovernanceVotes(s state.State) ([]*primitives.GovernanceVote, state.State) {
	p.governanceVoteLock.Lock()
	defer p.governanceVoteLock.Unlock()
	votes := make([]*primitives.GovernanceVote, 0, primitives.MaxGovernanceVotesPerBlock)
	newMempool := make(map[chainhash.Hash]*primitives.GovernanceVote)

	for k, gv := range p.governanceVotes {
		if err := s.ProcessGovernanceVote(gv); err != nil {
			continue
		}
		// if there is no error, it can be part of the new mempool
		newMempool[k] = gv

		if len(votes) < primitives.MaxGovernanceVotesPerBlock {
			votes = append(votes, gv)
		}
	}

	p.governanceVotes = newMempool

	return votes, s
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
func (p *pool) Start() error {
	p.Load()

	if err := p.host.RegisterTopicHandler(p2p.MsgVoteCmd, p.handleVote); err != nil {
		return err
	}

	if err := p.host.RegisterTopicHandler(p2p.MsgDepositCmd, p.handleDeposit); err != nil {
		return nil
	}

	if err := p.host.RegisterTopicHandler(p2p.MsgDepositsCmd, p.handleDeposits); err != nil {
		return nil
	}

	if err := p.host.RegisterTopicHandler(p2p.MsgExitCmd, p.handleExit); err != nil {
		return nil
	}

	if err := p.host.RegisterTopicHandler(p2p.MsgExitsCmd, p.handleExits); err != nil {
		return nil
	}

	if err := p.host.RegisterTopicHandler(p2p.MsgPartialExitsCmd, p.handlePartialExits); err != nil {
		return nil
	}

	if err := p.host.RegisterTopicHandler(p2p.MsgProofsCmd, p.handleProofs); err != nil {
		return nil
	}

	if err := p.host.RegisterTopicHandler(p2p.MsgTxCmd, p.handleTx); err != nil {
		return nil
	}

	if err := p.host.RegisterTopicHandler(p2p.MsgGovernanceCmd, p.handleGovernance); err != nil {
		return nil
	}

	if err := p.host.RegisterTopicHandler(p2p.MsgMultiSignatureTxCmd, p.handleMultiSignatureTx); err != nil {
		return nil
	}

	return nil
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
