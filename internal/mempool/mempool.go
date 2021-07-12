package mempool

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/VictoriaMetrics/fastcache"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/host"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"sort"
	"sync"
)

type Pool interface {
	Start()
	Close()

	AddVote(d *primitives.MultiValidatorVote, s state.State) error
	AddDeposit(d *primitives.Deposit) error
	AddExit(d *primitives.Exit) error
	AddPartialExit(d *primitives.PartialExit) error
	AddTx(d *primitives.Tx) error
	AddVoteSlashing(d *primitives.VoteSlashing) error
	AddProposerSlashing(d *primitives.ProposerSlashing) error
	AddRANDAOSlashing(d *primitives.RANDAOSlashing) error

	GetVotes(slotToPropose uint64, s state.State, index uint64) []*primitives.MultiValidatorVote
	GetDeposits(s state.State) ([]*primitives.Deposit, state.State)
	GetExits(s state.State) ([]*primitives.Exit, state.State)
	GetPartialExits(s state.State) ([]*primitives.PartialExit, state.State)
	GetTxs(s state.State, feeReceiver [20]byte) ([]*primitives.Tx, state.State)
	GetVoteSlashings(s state.State) ([]*primitives.VoteSlashing, state.State)
	GetProposerSlashings(s state.State) ([]*primitives.ProposerSlashing, state.State)
	GetRANDAOSlashings(s state.State) ([]*primitives.RANDAOSlashing, state.State)

	RemoveByBlock(b *primitives.Block, s state.State)
}

type pool struct {
	netParams *params.ChainParams
	log       logger.Logger
	ctx       context.Context

	chain chain.Blockchain
	host  host.Host
	//lastActionManager actionmanager.LastActionManager

	pool *fastcache.Cache

	singleVotes     map[[32]byte][]*primitives.MultiValidatorVote
	singleVotesLock sync.Mutex

	txKeys              sync.Map
	votesKeys           sync.Map
	depositKeys         sync.Map
	exitKeys            sync.Map
	partialExitKeys     sync.Map
	governanceVotesKeys sync.Map
	coinProofsKeys      sync.Map

	voteSlashings []*primitives.VoteSlashing

	proposerSlashings []*primitives.ProposerSlashing

	randaoSlashings []*primitives.RANDAOSlashing
}

func (p *pool) AddVote(d *primitives.MultiValidatorVote, s state.State) error {
	if err := s.IsVoteValid(d); err != nil {
		return err
	}

	voteData := d.Data
	voteHash := d.Data.Hash()

	firstSlotAllowedToInclude := d.Data.Slot + p.netParams.MinAttestationInclusionDelay

	currentState, err := p.chain.State().TipStateAtSlot(firstSlotAllowedToInclude)
	if err != nil {
		p.log.Error(err)
		return err
	}

	/*committee, err := currentState.GetVoteCommittee(d.Data.Slot)
	if err != nil {
		p.log.Error(err)
		return err
	}*/

	// Register voting action for validators included on the vote
	/*for i, c := range committee {
		if d.ParticipationBitfield.Get(uint(i)) {
			p.lastActionManager.RegisterAction(currentState.GetValidatorRegistry()[c].PubKey, time.Now(), d.Data.Nonce)
		}
	}*/

	// Slashing check
	// This check iterates over all the votes on the pool.
	// Checks if the new vote data matches any pool vote data hash.
	// If that check fails, we should check for validators submitting twice different votes.
	p.votesKeys.Range(func(key, value interface{}) bool {
		hash := key.(chainhash.Hash)
		cKey := appendKey(hash[:], PoolTypeVote)
		raw := p.pool.Get(nil, cKey)
		v := new(primitives.MultiValidatorVote)
		err := v.Unmarshal(raw)
		if err != nil {
			return true
		}
		if bytes.Equal(voteHash[:], hash[:]) {
			return true
		}
		if currentState.GetSlot() >= v.Data.LastSlotValid(p.netParams) {
			p.votesKeys.Delete(key)
			p.pool.Del(cKey)
			return true
		}

		var votingValidators = make(map[uint64]struct{})
		var intersect []uint64

		vote1Committee, err := currentState.GetVoteCommittee(v.Data.Slot)
		if err != nil {
			p.log.Error(err)
			return true
		}
		vote2Committee, err := currentState.GetVoteCommittee(d.Data.Slot)
		if err != nil {
			p.log.Error(err)
			return true
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
					return true
				}
				return true
			}
		}
		return true
	})

	// Check if vote is already on pool.
	// If a vote with same vote data is found we should check the signatures.
	// If the signatures are the same it means is a duplicated vote for network (probably a relayed vote).
	// If the signatures don't match, we should aggregate both signatures and merge the bitlists.
	// IMPORTANT: 	We should never allow a vote that conflicts a previous vote to be added to the pool.
	// 				That should be checked against all votes on pool comparing bitlists.
	key := appendKey(voteHash[:], PoolTypeVote)
	vRaw, ok := p.pool.HasGet(nil, key)
	if ok {
		v := new(primitives.MultiValidatorVote)
		err = v.Unmarshal(vRaw)
		if err != nil {
			return err
		}
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

			raw, err := newVote.Marshal()
			if err != nil {
				return err
			}
			p.pool.Set(key, raw)
			p.votesKeys.Store(voteHash, struct{}{})
			p.singleVotesLock.Lock()
			p.singleVotes[voteHash] = append(p.singleVotes[voteHash], d)
			p.singleVotesLock.Unlock()
		}
	} else {
		p.log.Debugf("adding vote to the mempool with %d votes", len(d.ParticipationBitfield.BitIndices()))
		raw, err := d.Marshal()
		if err != nil {
			return err
		}
		p.pool.Set(key, raw)
		p.votesKeys.Store(voteHash, struct{}{})
		p.singleVotesLock.Lock()
		p.singleVotes[voteHash] = []*primitives.MultiValidatorVote{d}
		p.singleVotesLock.Unlock()
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

	key := appendKey(d.Data.PublicKey[:], PoolTypeDeposit)

	ok := p.pool.Has(key)
	if !ok {
		p.pool.Set(key, raw)
		p.depositKeys.Store(d.Data.PublicKey, struct{}{})
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
		p.exitKeys.Store(d.ValidatorPubkey, struct{}{})
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
		p.partialExitKeys.Store(d.ValidatorPubkey, struct{}{})
	}

	return nil
}

func (p *pool) AddTx(d *primitives.Tx) error {

	cs := p.chain.State().TipState().GetCoinsState()

	fpkh, err := d.FromPubkeyHash()
	if err != nil {
		return err
	}

	// Check the state for a nonce lower than the used in transaction
	if stateNonce, ok := cs.Nonces[fpkh]; ok && d.Nonce < stateNonce || !ok && d.Nonce != 1 {
		return errors.New("invalid nonce against state")
	}

	if d.Fee < 5000 {
		return errors.New("transaction doesn't include enough fee")
	}

	if cs.Balances[fpkh] < d.Amount+d.Fee {
		return fmt.Errorf("insufficient balance of %d for %d transaction", cs.Balances[fpkh], d.Amount+d.Fee)
	}

	if err := d.VerifySig(); err != nil {
		return err
	}

	txKey := appendKeyWithNonce(fpkh, d.Nonce)
	key := appendKey(txKey[:], PoolTypeTx)
	ok := p.pool.Has(key)
	if !ok {
		raw, err := d.Marshal()
		if err != nil {
			return err
		}
		p.pool.Set(key, raw)
		p.txKeys.Store(txKey, struct{}{})
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

func (p *pool) GetVotes(slotToPropose uint64, s state.State, index uint64) []*primitives.MultiValidatorVote {

	var keys []chainhash.Hash
	p.votesKeys.Range(func(key, value interface{}) bool {
		pubKey := key.(chainhash.Hash)
		keys = append(keys, pubKey)
		if len(keys) >= primitives.MaxVotesPerBlock {
			return false
		}
		return true
	})

	var votes []*primitives.MultiValidatorVote

	for i := range keys {

		key := appendKey(keys[i][:], PoolTypeVote)

		raw := p.pool.Get(nil, key)

		d := new(primitives.MultiValidatorVote)

		err := d.Unmarshal(raw)
		if err != nil {
			p.pool.Del(key)
			p.depositKeys.Delete(keys[i])
			continue
		}

		if slotToPropose >= d.Data.FirstSlotValid(p.netParams) && slotToPropose <= d.Data.LastSlotValid(p.netParams) {
			err = s.ProcessVote(d, index)
			if err != nil {
				p.log.Error(err)
				p.pool.Del(key)
				p.depositKeys.Delete(keys[i])
				continue
			}
			votes = append(votes, d)
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
			return false
		}
		return true
	})

	var deposits []*primitives.Deposit
	for i := range keys {

		key := appendKey(keys[i][:], PoolTypeDeposit)

		raw := p.pool.Get(nil, key)

		d := new(primitives.Deposit)

		err := d.Unmarshal(raw)
		if err != nil {
			p.pool.Del(key)
			p.depositKeys.Delete(keys[i])
			continue
		}

		if err := s.ApplyDeposit(d); err != nil {
			p.pool.Del(key)
			p.depositKeys.Delete(keys[i])
			continue
		}

		deposits = append(deposits, d)
	}

	return deposits, s
}

func (p *pool) GetExits(s state.State) ([]*primitives.Exit, state.State) {

	var keys [][48]byte
	p.exitKeys.Range(func(key, value interface{}) bool {
		pubKey := key.([48]byte)
		keys = append(keys, pubKey)
		if len(keys) >= primitives.MaxExitsPerBlock {
			return false
		}
		return true
	})

	var exits []*primitives.Exit
	for i := range keys {

		key := appendKey(keys[i][:], PoolTypeExit)

		raw := p.pool.Get(nil, key)

		d := new(primitives.Exit)

		err := d.Unmarshal(raw)
		if err != nil {
			p.pool.Del(key)
			p.exitKeys.Delete(keys[i])
			continue
		}

		if err := s.ApplyExit(d); err != nil {
			p.pool.Del(key)
			p.exitKeys.Delete(keys[i])
			continue
		}

		exits = append(exits, d)
	}

	return exits, s
}

func (p *pool) GetPartialExits(s state.State) ([]*primitives.PartialExit, state.State) {

	var keys [][48]byte
	p.partialExitKeys.Range(func(key, value interface{}) bool {
		pubKey := key.([48]byte)
		keys = append(keys, pubKey)
		if len(keys) >= primitives.MaxPartialExitsPerBlock {
			return false
		}
		return true
	})

	var pexits []*primitives.PartialExit
	for i := range keys {

		key := appendKey(keys[i][:], PoolTypePartialExit)

		raw := p.pool.Get(nil, key)

		d := new(primitives.PartialExit)

		err := d.Unmarshal(raw)
		if err != nil {
			p.pool.Del(key)
			p.partialExitKeys.Delete(keys[i])
			continue
		}

		if err := s.ApplyPartialExit(d); err != nil {
			p.pool.Del(key)
			p.partialExitKeys.Delete(keys[i])
			continue
		}

		pexits = append(pexits, d)
	}

	return pexits, s
}

func (p *pool) GetTxs(s state.State, feeReceiver [20]byte) ([]*primitives.Tx, state.State) {

	var keys [][28]byte
	p.txKeys.Range(func(key, value interface{}) bool {
		pubKey := key.([28]byte)
		keys = append(keys, pubKey)
		if len(keys) >= primitives.MaxTxsPerBlock {
			return false
		}
		return true
	})

	var tempTxs []*primitives.Tx
	for i := range keys {

		key := appendKey(keys[i][:], PoolTypeTx)

		raw := p.pool.Get(nil, key)

		d := new(primitives.Tx)

		err := d.Unmarshal(raw)
		if err != nil {
			p.pool.Del(key)
			p.txKeys.Delete(keys[i])
			continue
		}

		tempTxs = append(tempTxs, d)
	}

	sort.Slice(tempTxs, func(i, j int) bool {
		return tempTxs[i].Nonce < tempTxs[j].Nonce
	})

	var txs []*primitives.Tx
	for _, tx := range tempTxs {
		hash := tx.Hash()
		key := appendKey(hash[:], PoolTypeTx)
		err := s.ApplyTransactionSingle(tx, feeReceiver)
		if err != nil {
			p.pool.Del(key)
			p.txKeys.Delete(key)
			continue
		}
		txs = append(txs, tx)
	}

	return txs, s
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

func (p *pool) RemoveByBlock(b *primitives.Block, s state.State) {

	// Check for votes on the block and remove them
	p.singleVotesLock.Lock()
	defer p.singleVotesLock.Unlock()
	for _, blockVote := range b.Votes {
		hash := blockVote.Data.Hash()
		key := appendKey(hash[:], PoolTypeVote)

		// If the vote is on pool and included on the block, remove it.
		rawVote, ok := p.pool.HasGet(nil, key)
		if ok {

			p.pool.Del(key)
			p.votesKeys.Delete(hash)
			poolVote := new(primitives.MultiValidatorVote)
			err := poolVote.Unmarshal(rawVote)
			if err != nil {
				continue
			}
			// If the mempool vote participation is greater than votes included on block we check the poolIndividuals
			// if there are more votes on the poolIndividuals that were not included on the block, we aggregate them and
			// add a new vote to the mempool.
			// including the missing votes.
			if len(poolVote.ParticipationBitfield.BitIndices()) > len(blockVote.ParticipationBitfield.BitIndices()) {
				p.log.Debug("incomplete vote submission detected aggregating and constructing missing vote")

				individuals := p.singleVotes[hash]
				delete(p.singleVotes, hash)

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

				if len(sigs) > 0 {
					aggSig := bls.AggregateSignatures(sigs)
					var voteSig [96]byte
					copy(voteSig[:], aggSig.Marshal())

					newVote := &primitives.MultiValidatorVote{
						Data:                  poolVote.Data,
						ParticipationBitfield: newBitfield,
						Sig:                   voteSig,
					}

					raw, err := newVote.Marshal()
					if err != nil {
						continue
					}
					p.pool.Set(key, raw)
					p.votesKeys.Store(hash, struct{}{})
				}

			} else {
				delete(p.singleVotes, hash)
			}
		}
	}

	for _, d := range b.Deposits {
		key := appendKey(d.Data.PublicKey[:], PoolTypeDeposit)
		ok := p.pool.Has(key)
		if ok {
			p.pool.Del(key)
			p.depositKeys.Delete(d.Data.PublicKey)
		}
	}

	for _, e := range b.Exits {
		key := appendKey(e.ValidatorPubkey[:], PoolTypeExit)
		ok := p.pool.Has(key)
		if ok {
			p.pool.Del(key)
			p.exitKeys.Delete(e.ValidatorPubkey)
		}
	}

	for _, pe := range b.PartialExit {
		key := appendKey(pe.ValidatorPubkey[:], PoolTypePartialExit)
		ok := p.pool.Has(key)
		if ok {
			p.pool.Del(key)
			p.partialExitKeys.Delete(pe.ValidatorPubkey)
		}
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

	for _, tx := range b.Txs {
		fpkh, err := tx.FromPubkeyHash()
		if err != nil {
			continue
		}

		txKey := appendKeyWithNonce(fpkh, tx.Nonce)
		key := appendKey(txKey[:], PoolTypeTx)
		ok := p.pool.Has(key)
		if ok {
			p.pool.Del(key)
			p.txKeys.Delete(txKey)
		}

	}

}

var _ Pool = &pool{}

func (p *pool) Close() {
	datapath := config.GlobalFlags.DataPath
	_ = p.pool.SaveToFile(datapath + "/mempool")
}

// Start initializes the pool listeners
func (p *pool) Start() {

	p.host.RegisterTopicHandler(p2p.MsgVoteCmd, p.handleVote)

	p.host.RegisterTopicHandler(p2p.MsgDepositsCmd, p.handleDeposits)

	p.host.RegisterTopicHandler(p2p.MsgExitsCmd, p.handleExits)

	p.host.RegisterTopicHandler(p2p.MsgPartialExitsCmd, p.handlePartialExits)

	p.host.RegisterTopicHandler(p2p.MsgTxCmd, p.handleTx)

	return

}

func NewPool(ch chain.Blockchain, h host.Host /*, manager actionmanager.LastActionManager*/) Pool {
	datapath := config.GlobalFlags.DataPath

	var cache *fastcache.Cache
	var err error
	cache, err = fastcache.LoadFromFile(datapath + "/mempool")
	if err != nil {
		cache = fastcache.New(300 * 1024 * 1024)
	}
	return &pool{
		netParams: config.GlobalParams.NetParams,
		log:       config.GlobalParams.Logger,
		ctx:       config.GlobalParams.Context,
		chain:     ch,
		host:      h,
		//lastActionManager: manager,

		pool: cache,

		singleVotes:       make(map[[32]byte][]*primitives.MultiValidatorVote),
		voteSlashings:     []*primitives.VoteSlashing{},
		proposerSlashings: []*primitives.ProposerSlashing{},
		randaoSlashings:   []*primitives.RANDAOSlashing{},
	}
}
