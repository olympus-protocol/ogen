package mempool

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"github.com/VictoriaMetrics/fastcache"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"sort"
	"sync"
)

var (
	ErrorAccountNotOnMempool = errors.New("account not on pool")
)

type Pool interface {
	Start() error

	AddVote(d *primitives.MultiValidatorVote, s state.State) error
	AddDeposit(d *primitives.Deposit) error
	AddExit(d *primitives.Exit) error
	AddPartialExit(d *primitives.PartialExit) error
	AddTx(d *primitives.Tx) error
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

	RemoveByBlock(b *primitives.Block, s state.State)
}

type pool struct {
	netParams *params.ChainParams
	log       logger.Logger
	ctx       context.Context

	chain             chain.Blockchain
	host              hostnode.HostNode
	lastActionManager actionmanager.LastActionManager

	pool *fastcache.Cache

	votesLock sync.Mutex
	votes     map[chainhash.Hash]*primitives.MultiValidatorVote

	intidivualVotes     map[chainhash.Hash][]*primitives.MultiValidatorVote
	intidivualVotesLock sync.Mutex

	depositKeys         sync.Map
	exitKeys            sync.Map
	partialExitKeys     sync.Map
	governanceVotesKeys sync.Map
	coinProofsKeys      sync.Map

	txsLock sync.Mutex
	txs     map[[20]byte]*txItem

	voteSlashings []*primitives.VoteSlashing

	proposerSlashings []*primitives.ProposerSlashing

	randaoSlashings []*primitives.RANDAOSlashing
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

	committee, err := currentState.GetVoteCommittee(d.Data.Slot)
	if err != nil {
		p.log.Error(err)
		return err
	}

	// Register voting action for validators included on the vote
	for i, c := range committee {
		if d.ParticipationBitfield.Get(uint(i)) {
			p.lastActionManager.RegisterAction(currentState.GetValidatorRegistry()[c].PubKey, d.Data.Nonce)
		}
	}

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

	raw, err := d.Marshal()
	if err != nil {
		return err
	}

	key := appendKey(d.PublicKey[:], PoolTypeDeposit)

	ok := p.pool.Has(key)
	if !ok {
		p.pool.Set(key, raw)
		p.depositKeys.Store(d.PublicKey[:], struct{}{})
	}

	return nil
}

func (p *pool) AddExit(d *primitives.Exit) error {
	s := p.chain.State().TipState()

	if err := s.IsExitValid(d); err != nil {
		return err
	}

	raw, err := d.Marshal()
	if err != nil {
		return err
	}

	key := appendKey(d.ValidatorPubkey[:], PoolTypeExit)

	ok := p.pool.Has(key)
	if !ok {
		p.pool.Set(key, raw)
		p.exitKeys.Store(d.ValidatorPubkey[:], struct{}{})
	}

	return nil
}

func (p *pool) AddPartialExit(d *primitives.PartialExit) error {
	s := p.chain.State().TipState()

	if err := s.IsPartialExitValid(d); err != nil {
		return err
	}

	raw, err := d.Marshal()
	if err != nil {
		return err
	}

	key := appendKey(d.ValidatorPubkey[:], PoolTypePartialExit)

	ok := p.pool.Has(key)
	if !ok {
		p.pool.Set(key, raw)
		p.partialExitKeys.Store(d.ValidatorPubkey[:], struct{}{})
	}

	return nil
}

func (p *pool) AddTx(d *primitives.Tx) error {
	p.txsLock.Lock()
	defer p.txsLock.Unlock()

	cs := p.chain.State().TipState().GetCoinsState()

	fpkh, err := d.FromPubkeyHash()
	if err != nil {
		return err
	}

	if lnb, ok := p.pool.HasGet(nil, appendKey(fpkh[:], PoolTypeLatestNonce)); ok {
		latestNonce := binary.LittleEndian.Uint64(lnb)
		if d.Nonce < latestNonce {
			return errors.New("invalid nonce against pool map")
		}
	}

	// Check the state for a nonce lower than the used in transaction
	if stateNonce, ok := cs.Nonces[fpkh]; ok && d.Nonce < stateNonce || !ok && d.Nonce != 1 {
		return errors.New("invalid nonce against state")
	}

	if d.Fee < 5000 {
		return errors.New("transaction doesn't include enough fee")
	}

	mpi, ok := p.txs[fpkh]

	if !ok {
		p.txs[fpkh] = newCoinMempoolItem()
		mpi = p.txs[fpkh]
		if err := mpi.add(d, cs.Balances[fpkh]); err != nil {
			return err
		}
		p.pool.Set(appendKey(fpkh[:], PoolTypeLatestNonce), nonceToBytes(d.Nonce))
	} else {
		if err := mpi.add(d, cs.Balances[fpkh]); err != nil {
			return err
		}
		p.pool.Set(appendKey(fpkh[:], PoolTypeLatestNonce), nonceToBytes(d.Nonce))
	}

	return nil
}

func (p *pool) AddVoteSlashing(d *primitives.VoteSlashing) error {
	p.log.Warn("WARNING: Vote slashing condition detected.")

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

	voteHash := d.Hash()

	raw, err := d.Marshal()
	if err != nil {
		return err
	}

	key := appendKey(voteHash[:], PoolTypeGovernanceVote)
	ok := p.pool.Has(key)
	if !ok {
		p.pool.Set(key, raw)
		p.governanceVotesKeys.Store(key, struct{}{})
	}

	return nil
}

func (p *pool) AddCoinProof(d *burnproof.CoinsProofSerializable) error {
	s := p.chain.State().TipState()

	if err := s.IsCoinProofValid(d); err != nil {
		return err
	}

	hash := d.Hash()
	raw, err := d.Marshal()
	if err != nil {
		return err
	}

	key := appendKey(hash[:], PoolTypeCoinProof)
	ok := p.pool.Has(key)
	if !ok {
		p.pool.Set(key, raw)
		p.coinProofsKeys.Store(key, struct{}{})
	}
	return nil
}

func (p *pool) GetAccountNonce(pkh [20]byte) (uint64, error) {
	nB, ok := p.pool.HasGet(nil, appendKey(pkh[:], PoolTypeLatestNonce))
	if !ok {
		return 0, ErrorAccountNotOnMempool
	}

	nonce := binary.LittleEndian.Uint64(nB)

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

	var keys [][48]byte
	p.depositKeys.Range(func(key, value interface{}) bool {
		pubKey := key.([48]byte)
		keys = append(keys, pubKey)
		if len(keys) >= primitives.MaxDepositsPerBlock {
			return true
		}
		return true
	})

	deposits := make([]*primitives.Deposit, len(keys))
	for i := range deposits {

		key := appendKey(keys[i][:], PoolTypeDeposit)

		raw := p.pool.Get(nil, key)

		d := new(primitives.Deposit)

		err := d.Unmarshal(raw)
		if err != nil {
			p.pool.Del(key)
			p.depositKeys.Delete(keys[i][:])
			continue
		}

		deposits[i] = d
	}

	return deposits, s
}

func (p *pool) GetExits(s state.State) ([]*primitives.Exit, state.State) {

	var keys [][48]byte
	p.exitKeys.Range(func(key, value interface{}) bool {
		pubKey := key.([48]byte)
		keys = append(keys, pubKey)
		if len(keys) >= primitives.MaxExitsPerBlock {
			return true
		}
		return true
	})

	exits := make([]*primitives.Exit, len(keys))
	for i := range exits {
		key := appendKey(keys[i][:], PoolTypeExit)

		raw := p.pool.Get(nil, key)

		d := new(primitives.Exit)

		err := d.Unmarshal(raw)
		if err != nil {
			p.pool.Del(key)
			p.exitKeys.Delete(keys[i][:])
			continue
		}

		exits[i] = d
	}

	return exits, s
}

func (p *pool) GetPartialExits(s state.State) ([]*primitives.PartialExit, state.State) {

	var keys [][48]byte
	p.partialExitKeys.Range(func(key, value interface{}) bool {
		pubKey := key.([48]byte)
		keys = append(keys, pubKey)
		if len(keys) >= primitives.MaxPartialExitsPerBlock {
			return true
		}
		return true
	})

	pexits := make([]*primitives.PartialExit, len(keys))
	for i := range pexits {
		key := appendKey(keys[i][:], PoolTypePartialExit)

		raw := p.pool.Get(nil, key)

		d := new(primitives.PartialExit)

		err := d.Unmarshal(raw)
		if err != nil {
			p.pool.Del(key)
			p.partialExitKeys.Delete(keys[i][:])
			continue
		}

		pexits[i] = d
	}

	return pexits, s
}

func (p *pool) GetCoinProofs(s state.State) ([]*burnproof.CoinsProofSerializable, state.State) {

	var keys [][32]byte
	p.coinProofsKeys.Range(func(key, value interface{}) bool {
		pubKey := key.([32]byte)
		keys = append(keys, pubKey)
		if len(keys) >= primitives.MaxCoinProofsPerBlock {
			return true
		}
		return true
	})

	coinProofs := make([]*burnproof.CoinsProofSerializable, len(keys))
	for i := range coinProofs {
		key := appendKey(keys[i][:], PoolTypeCoinProof)

		raw := p.pool.Get(nil, key)

		d := new(burnproof.CoinsProofSerializable)

		err := d.Unmarshal(raw)
		if err != nil {
			p.pool.Del(key)
			p.coinProofsKeys.Delete(keys[i][:])
			continue
		}

		coinProofs[i] = d
	}

	return coinProofs, s

}

func (p *pool) GetTxs(s state.State, feeReceiver [20]byte) ([]*primitives.Tx, state.State) {
	p.txsLock.Lock()
	defer p.txsLock.Unlock()

	allTransactions := make([]*primitives.Tx, 0, primitives.MaxTxsPerBlock)

	for _, addr := range p.txs {
		nonces := make([]int, 0, len(addr.transactions))
		for k := range addr.transactions {
			nonces = append(nonces, int(k))
		}

		sort.Ints(nonces)

		for _, nonce := range nonces {
			tx := addr.transactions[uint64(nonce)]
			if err := s.ApplyTransactionSingle(tx, feeReceiver); err != nil {
				continue
			}
			if len(allTransactions) < primitives.MaxTxsPerBlock {
				allTransactions = append(allTransactions, tx)
			}
		}

	}

	return allTransactions, s
}

func (p *pool) GetVoteSlashings(s state.State) ([]*primitives.VoteSlashing, state.State) {

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

	var keys [][32]byte
	p.coinProofsKeys.Range(func(key, value interface{}) bool {
		pubKey := key.([32]byte)
		keys = append(keys, pubKey)
		if len(keys) >= primitives.MaxGovernanceVotesPerBlock {
			return true
		}
		return true
	})

	governanceVotes := make([]*primitives.GovernanceVote, len(keys))
	for i := range governanceVotes {
		key := appendKey(keys[i][:], PoolTypeGovernanceVote)

		raw := p.pool.Get(nil, key)

		d := new(primitives.GovernanceVote)

		err := d.Unmarshal(raw)
		if err != nil {
			p.pool.Del(key)
			p.governanceVotesKeys.Delete(keys[i][:])
			continue
		}

		governanceVotes[i] = d
	}

	return governanceVotes, s
}

func (p *pool) RemoveByBlock(b *primitives.Block, s state.State) {
	netParams := config.GlobalParams.NetParams
	p.votesLock.Lock()
	p.intidivualVotesLock.Lock()

	for _, v := range p.votes {
		voteHash := v.Data.Hash()
		if b.Header.Slot >= v.Data.LastSlotValid(netParams) {
			delete(p.votes, voteHash)
			delete(p.intidivualVotes, voteHash)
		}
	}

	// Check for votes on the block and remove them
	for _, blockVote := range b.Votes {
		voteHash := blockVote.Data.Hash()

		// If the vote is on pool and included on the block, remove it.
		poolVote, ok := p.votes[voteHash]
		if ok {
			delete(p.votes, voteHash)

			// If the mempool vote participation is greater than votes included on block we check the poolIndividuals
			// if there are more votes on the poolIndividuals that were not included on the block, we aggregate them and
			// add a new vote to the mempool.
			// including the missing votes.
			if len(poolVote.ParticipationBitfield.BitIndices()) > len(blockVote.ParticipationBitfield.BitIndices()) {
				p.log.Debug("incomplete vote submission detected aggregating and constructing missing vote")
				individuals := p.intidivualVotes[voteHash]
				// First we extract the included vote for the individuals slice
				var votesToAggregate []*primitives.MultiValidatorVote
				for _, iv := range individuals {
					intersect := iv.ParticipationBitfield.Intersect(blockVote.ParticipationBitfield)
					if len(intersect) == 0 {
						votesToAggregate = append(votesToAggregate, iv)
					}
				}
				p.log.Debugf("found %d individual votes not included", len(votesToAggregate))

				newBitfield := bitfield.NewBitlist(poolVote.ParticipationBitfield.Len())

				var sigs []common.Signature
				for _, missingVote := range votesToAggregate {
					sig, err := missingVote.Signature()
					if err != nil {
						return
					}
					sigs = append(sigs, sig)

					for _, idx := range missingVote.ParticipationBitfield.BitIndices() {
						newBitfield.Set(uint(idx))
					}
				}

				aggSig := bls.AggregateSignatures(sigs)
				var voteSig [96]byte
				copy(voteSig[:], aggSig.Marshal())

				newVote := &primitives.MultiValidatorVote{
					Data:                  poolVote.Data,
					ParticipationBitfield: newBitfield,
					Sig:                   voteSig,
				}

				p.votes[voteHash] = newVote
			}
		}
	}

	p.log.Debugf("tracking %d aggregated votes and %d individual votes in vote mempool", len(p.votes), len(p.intidivualVotes))

	p.votesLock.Unlock()
	p.intidivualVotesLock.Unlock()

	for _, d := range b.Deposits {
		key := appendKey(d.PublicKey[:], PoolTypeDeposit)
		p.pool.Del(key)
		p.depositKeys.Delete(d.PublicKey[:])
	}

	for _, e := range b.Exits {
		key := appendKey(e.ValidatorPubkey[:], PoolTypeExit)
		p.pool.Del(key)
		p.exitKeys.Delete(e.ValidatorPubkey[:])
	}

	for _, pe := range b.PartialExit {
		key := appendKey(pe.ValidatorPubkey[:], PoolTypePartialExit)
		p.pool.Del(key)
		p.partialExitKeys.Delete(pe.ValidatorPubkey[:])
	}

	for _, gv := range b.GovernanceVotes {
		hash := gv.Hash()
		key := appendKey(hash[:], PoolTypeGovernanceVote)
		p.pool.Del(key)
		p.partialExitKeys.Delete(hash[:])
	}

	for _, c := range b.CoinProofs {
		hash := c.Hash()

		key := appendKey(hash[:], PoolTypeCoinProof)
		p.pool.Del(key)
		p.partialExitKeys.Delete(hash[:])
	}

	newProposerSlashings := make([]*primitives.ProposerSlashing, 0, len(p.proposerSlashings))
	for _, ps := range p.proposerSlashings {
		psHash := ps.Hash()
		if b.Header.Slot >= ps.BlockHeader2.Slot+p.netParams.EpochLength-1 {
			continue
		}

		if b.Header.Slot >= ps.BlockHeader1.Slot+p.netParams.EpochLength-1 {
			continue
		}

		for _, blockSlashing := range b.ProposerSlashings {
			blockSlashingHash := blockSlashing.Hash()

			if blockSlashingHash.IsEqual(&psHash) {
				continue
			}
		}

		if _, err := s.IsProposerSlashingValid(ps); err != nil {
			continue
		}

		newProposerSlashings = append(newProposerSlashings, ps)
	}
	p.proposerSlashings = newProposerSlashings

	newVoteSlashings := make([]*primitives.VoteSlashing, 0, len(p.voteSlashings))
	for _, vs := range p.voteSlashings {
		vsHash := vs.Hash()
		if b.Header.Slot >= vs.Vote1.Data.LastSlotValid(p.netParams) {
			continue
		}

		if b.Header.Slot >= vs.Vote2.Data.LastSlotValid(p.netParams) {
			continue
		}

		for _, voteSlashing := range b.VoteSlashings {
			voteSlashingHash := voteSlashing.Hash()

			if voteSlashingHash.IsEqual(&vsHash) {
				continue
			}
		}

		if _, err := s.IsVoteSlashingValid(vs); err != nil {
			continue
		}

		newVoteSlashings = append(newVoteSlashings, vs)
	}
	p.voteSlashings = newVoteSlashings

	newRANDAOSlashings := make([]*primitives.RANDAOSlashing, 0, len(p.randaoSlashings))
	for _, rs := range p.randaoSlashings {
		rsHash := rs.Hash()

		for _, blockSlashing := range b.VoteSlashings {
			blockSlashingHash := blockSlashing.Hash()

			if blockSlashingHash.IsEqual(&rsHash) {
				continue
			}
		}

		if _, err := s.IsRANDAOSlashingValid(rs); err != nil {
			continue
		}

		newRANDAOSlashings = append(newRANDAOSlashings, rs)
	}
	p.randaoSlashings = newRANDAOSlashings

	p.txsLock.Lock()

	for _, tx := range b.Txs {
		fpkh, err := tx.FromPubkeyHash()
		if err != nil {
			continue
		}

		it, found := p.txs[fpkh]
		if !found {
			continue
		}
		it.removeBefore(tx.Nonce)
		if it.balanceSpent == 0 {
			delete(p.txs, fpkh)
		}
		key := appendKey(fpkh[:], PoolTypeLatestNonce)
		if lnb, ok := p.pool.HasGet(nil, key); ok {
			nonce := binary.LittleEndian.Uint64(lnb)
			if tx.Nonce == nonce {
				p.pool.Del(key)
			}
		}
	}

	p.txsLock.Unlock()

}

var _ Pool = &pool{}

// Start initializes the pool listeners
func (p *pool) Start() error {

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

	return nil
}

func NewPool(ch chain.Blockchain, hostnode hostnode.HostNode, manager actionmanager.LastActionManager) Pool {
	return &pool{
		netParams:         config.GlobalParams.NetParams,
		log:               config.GlobalParams.Logger,
		ctx:               config.GlobalParams.Context,
		chain:             ch,
		host:              hostnode,
		lastActionManager: manager,

		pool: fastcache.New(300 * 1024 * 1024),

		votes:           make(map[chainhash.Hash]*primitives.MultiValidatorVote),
		intidivualVotes: make(map[chainhash.Hash][]*primitives.MultiValidatorVote),
		txs:             make(map[[20]byte]*txItem),

		voteSlashings:     []*primitives.VoteSlashing{},
		proposerSlashings: []*primitives.ProposerSlashing{},
		randaoSlashings:   []*primitives.RANDAOSlashing{},
	}
}
