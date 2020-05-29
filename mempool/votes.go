package mempool

import (
	"bytes"
	"context"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/logger"
)

type mempoolVote struct {
	individualVotes         []*primitives.SingleValidatorVote
	participationBitfield   bls.Bitfield
	participatingValidators map[uint32]struct{}
	voteData                *primitives.VoteData
}

func (mv *mempoolVote) getVoteByOffset(offset uint32) (*primitives.SingleValidatorVote, bool) {
	if !mv.participationBitfield.Get(uint(offset)) {
		return nil, false
	}

	var vote *primitives.SingleValidatorVote
	for _, v := range mv.individualVotes {
		if v.Offset == offset {
			vote = v
		}
	}

	return vote, vote != nil
}

func (mv *mempoolVote) add(vote *primitives.SingleValidatorVote, voter uint32) {
	if mv.participationBitfield.Get(uint(vote.Offset)) {
		return
	}
	mv.individualVotes = append(mv.individualVotes, vote)
	mv.participationBitfield.Set(uint(vote.Offset))
	mv.participatingValidators[voter] = struct{}{}
}

func (mv *mempoolVote) remove(participationBitfield []uint8) (shouldRemove bool) {
	shouldRemove = true
	newVotes := make([]*primitives.SingleValidatorVote, 0, len(mv.individualVotes))
	for _, v := range mv.individualVotes {
		if len(participationBitfield) >= int(v.Offset/8) {
			return shouldRemove
		}
		if participationBitfield[v.Offset/8]&(1<<uint(v.Offset%8)) == 0 {
			newVotes = append(newVotes, v)
			shouldRemove = false
		}
	}

	mv.individualVotes = newVotes
	for i, p := range participationBitfield {
		mv.participationBitfield[i] &= ^p
	}
	return shouldRemove
}

func newMempoolVote(outOf uint32, voteData *primitives.VoteData) *mempoolVote {
	return &mempoolVote{
		participationBitfield:   make([]uint8, (outOf+7)/8),
		individualVotes:         make([]*primitives.SingleValidatorVote, 0, outOf),
		voteData:                voteData,
		participatingValidators: make(map[uint32]struct{}),
	}
}

// VoteMempool is a mempool that keeps track of votes.
type VoteMempool struct {
	poolLock sync.Mutex
	pool     map[chainhash.Hash]*mempoolVote

	// index 0 is highest priorized, 1 is less, etc
	poolOrder []chainhash.Hash

	params     *params.ChainParams
	log        *logger.Logger
	ctx        context.Context
	blockchain *chain.Blockchain

	notifees     []VoteSlashingNotifee
	notifeesLock sync.Mutex
}

func shuffleVotes(vals []primitives.SingleValidatorVote) []primitives.SingleValidatorVote {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ret := make([]primitives.SingleValidatorVote, len(vals))
	perm := r.Perm(len(vals))
	for i, randIndex := range perm {
		ret[i] = vals[randIndex]
	}
	return ret
}

// func PickPercentVotes(vs []primitives.SingleValidatorVote, pct float32) []primitives.SingleValidatorVote {
// 	num := int(pct * float32(len(vs)))
// 	shuffledVotes := shuffleVotes(vs)
// 	return shuffledVotes[:num]
// }

// func (m *VoteMempool) GetVotesNotInBloom(bloom *bloom.BloomFilter) []primitives.SingleValidatorVote {
// 	votes := make([]primitives.SingleValidatorVote, 0)
// 	for _, vs := range m.pool {
// 		for _, v := range vs.individualVotes {
// 			vh := v.Hash()
// 			if bloom.Has(vh) {
// 				continue
// 			}

// 			votes = append(votes, *v)
// 		}
// 	}
// 	return votes
// }

// AddValidate validates, then adds the vote to the mempool.
func (m *VoteMempool) AddValidate(vote *primitives.SingleValidatorVote, state *primitives.State) error {
	if err := state.IsVoteValid(vote.AsMulti(), m.params); err != nil {
		return err
	}

	m.Add(vote)
	return nil
}

// sortMempool sorts the poolOrder so that the highest priority transactions come first
// and assumes you hold the poolLock.
func (m *VoteMempool) sortMempool() {
	sort.Slice(m.poolOrder, func(i, j int) bool {
		// return if i is higher priority than j
		iHash := m.poolOrder[i]
		jHash := m.poolOrder[j]
		iData := m.pool[iHash].voteData
		iNumVotes := len(m.pool[iHash].individualVotes)
		jData := m.pool[jHash].voteData
		jNumVotes := len(m.pool[iHash].individualVotes)

		// first sort by slot
		if iData.Slot < jData.Slot {
			return true
		} else if iData.Slot > jData.Slot {
			return false
		}

		if iNumVotes > jNumVotes {
			return true
		} else if iNumVotes < jNumVotes {
			return false
		}

		// arbitrary
		return true
	})
}

// Add adds a vote to the mempool.
func (m *VoteMempool) Add(vote *primitives.SingleValidatorVote) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	voteHash := vote.Data.Hash()

	firstSlotAllowedToInclude := vote.Data.Slot + m.params.MinAttestationInclusionDelay
	tipHash := m.blockchain.State().Tip().Hash
	view, err := m.blockchain.State().GetSubView(tipHash)
	if err != nil {
		m.log.Warnf("could not get block view representing current tip: %s", err)
		return
	}
	currentState, _, err := m.blockchain.State().GetStateForHashAtSlot(tipHash, firstSlotAllowedToInclude, &view, m.params)
	if err != nil {
		m.log.Warnf("error updating chain to attestation inclusion slot: %s", err)
		return
	}

	committee, err := currentState.GetVoteCommittee(vote.Data.Slot, m.params)
	if err != nil {
		m.log.Error(err)
		return
	}

	if vote.Offset >= uint32(len(committee)) {
		return
	}

	voter := committee[vote.Offset]

	// slashing check... check if this vote interferes with any
	// votes in the mempool
	voteData := vote.Data
	for h, v := range m.pool {
		if voteHash.IsEqual(&h) {
			continue
		}
		if _, ok := v.participatingValidators[voter]; ok {
			if v.voteData.IsDoubleVote(voteData) || v.voteData.IsSurroundVote(voteData) {
				if v.voteData.IsDoubleVote(voteData) {
					m.log.Warnf("found double vote for validator %d in vote %s and %s, reporting...", voter, vote.Data.String(), v.voteData.String())
				}
				if v.voteData.IsSurroundVote(voteData) {
					m.log.Warnf("found surround vote for validator %d in vote %s and %s, reporting...", voter, vote.Data.String(), v.voteData.String())
				}
				conflicting, found := v.getVoteByOffset(vote.Offset)
				if !found {
					return
				}
				for _, n := range m.notifees {
					n.NotifyIllegalVotes(vote.AsMulti(), conflicting.AsMulti())
				}
			}
			return
		}
	}

	if vs, found := m.pool[voteHash]; found {
		vs.add(vote, voter)
	} else {
		m.pool[voteHash] = newMempoolVote(vote.OutOf, &vote.Data)
		m.poolOrder = append(m.poolOrder, voteHash)
		m.pool[voteHash].add(vote, voter)
	}

	m.sortMempool()
}

// Get gets a vote from the mempool.
func (m *VoteMempool) Get(slot uint64, p *params.ChainParams) []primitives.MultiValidatorVote {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	votes := make([]primitives.MultiValidatorVote, 0)
	for _, i := range m.poolOrder {
		v := m.pool[i]
		if v.voteData.Slot < slot-p.MinAttestationInclusionDelay && slot <= v.voteData.Slot+m.params.EpochLength-1 {
			sigs := make([]*bls.Signature, 0)
			for _, v := range m.pool[i].individualVotes {
				sigs = append(sigs, &v.Signature)
			}
			sig := bls.AggregateSignatures(sigs)
			vote := primitives.MultiValidatorVote{
				Data:                  *m.pool[i].voteData,
				Signature:             *sig,
				ParticipationBitfield: append([]uint8(nil), v.participationBitfield...),
			}
			votes = append(votes, vote)
		}
	}

	return votes
}

// Assumes you already hold the poolLock mutex.
func (m *VoteMempool) removeFromOrder(h chainhash.Hash) {
	newOrder := make([]chainhash.Hash, 0, len(m.poolOrder)-1)

	for _, vh := range m.poolOrder {
		if !vh.IsEqual(&h) {
			newOrder = append(newOrder, vh)
		}
	}

	m.poolOrder = newOrder
}

// Remove removes mempool items that are no longer relevant.
func (m *VoteMempool) Remove(b *primitives.Block) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	for _, v := range b.Votes {
		voteHash := v.Data.Hash()

		var shouldRemove bool
		if vote, found := m.pool[voteHash]; found {
			shouldRemove = vote.remove(v.ParticipationBitfield)
		}
		if shouldRemove {
			delete(m.pool, voteHash)
			m.removeFromOrder(voteHash)
		}

		if b.Header.Slot >= v.Data.Slot+m.params.EpochLength-1 {
			delete(m.pool, voteHash)
			m.removeFromOrder(voteHash)
		}
	}
}

func (m *VoteMempool) handleSubscription(topic *pubsub.Subscription, id peer.ID) {
	for {
		msg, err := topic.Next(m.ctx)
		if err != nil {
			m.log.Warnf("error getting next message in votes topic: %s", err)
			return
		}

		if msg.GetFrom() == id {
			continue
		}

		txBuf := bytes.NewReader(msg.Data)
		tx := new(primitives.SingleValidatorVote)

		if err := tx.Decode(txBuf); err != nil {
			// TODO: ban peer
			m.log.Warnf("peer sent invalid vote: %s", err)
			continue
		}

		firstSlotAllowedToInclude := tx.Data.Slot + m.params.MinAttestationInclusionDelay
		tipHash := m.blockchain.State().Tip().Hash
		view, err := m.blockchain.State().GetSubView(tipHash)
		if err != nil {
			m.log.Warnf("could not get block view representing current tip: %s", err)
			continue
		}
		currentState, _, err := m.blockchain.State().GetStateForHashAtSlot(tipHash, firstSlotAllowedToInclude, &view, m.params)
		if err != nil {
			m.log.Warnf("error updating chain to attestation inclusion slot: %s", err)
			continue
		}

		err = m.AddValidate(tx, currentState)
		if err != nil {
			m.log.Warnf("error adding transaction to mempool: %s", err)
		}
	}
}

// Notify registers a notifee to be notified when illegal votes occur.
func (m *VoteMempool) Notify(notifee VoteSlashingNotifee) {
	m.notifeesLock.Lock()
	defer m.notifeesLock.Unlock()
	m.notifees = append(m.notifees, notifee)
}

// NewVoteMempool creates a new mempool.
func NewVoteMempool(ctx context.Context, log *logger.Logger, p *params.ChainParams, ch *chain.Blockchain, hostnode *peers.HostNode) (*VoteMempool, error) {
	vm := &VoteMempool{
		pool:       make(map[chainhash.Hash]*mempoolVote),
		params:     p,
		log:        log,
		ctx:        ctx,
		blockchain: ch,
		notifees:   make([]VoteSlashingNotifee, 0),
	}
	voteTopic, err := hostnode.Topic("votes")
	if err != nil {
		return nil, err
	}

	voteSub, err := voteTopic.Subscribe()
	if err != nil {
		return nil, err
	}

	go vm.handleSubscription(voteSub, hostnode.GetHost().ID())

	return vm, nil
}

// VoteSlashingNotifee is notified when an illegal vote occurs.
type VoteSlashingNotifee interface {
	NotifyIllegalVotes(vote1 *primitives.MultiValidatorVote, vote2 *primitives.MultiValidatorVote)
}
