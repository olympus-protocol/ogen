package peers

import (
	"math/rand"
	"sync"
	"time"

	"github.com/olympus-protocol/ogen/bloom"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type mempoolVote struct {
	individualVotes       []*primitives.SingleValidatorVote
	participationBitfield []uint8
	aggregateSignature    *bls.Signature
	voteData              *primitives.VoteData
}

func (mv *mempoolVote) add(vote *primitives.SingleValidatorVote) {
	if mv.participationBitfield[vote.Offset/8]&(1<<uint(vote.Offset%8)) > 0 {
		return
	}
	mv.individualVotes = append(mv.individualVotes, vote)
	mv.participationBitfield[vote.Offset/8] |= (1 << uint(vote.Offset%8))
	mv.aggregateSignature.AggregateSig(&vote.Signature)
}

func (mv *mempoolVote) remove(participationBitfield []uint8) (shouldRemove bool) {
	shouldRemove = true
	newVotes := make([]*primitives.SingleValidatorVote, 0, len(mv.individualVotes))
	for _, v := range mv.individualVotes {
		if participationBitfield[v.Offset/8]&(1<<uint(v.Offset%8)) == 0 {
			newVotes = append(newVotes, v)
			shouldRemove = false
		}
	}

	mv.individualVotes = newVotes
	for i, p := range participationBitfield {
		mv.participationBitfield[i] &= ^p
	}

	newAggSig := bls.NewAggregateSignature()
	for _, v := range newVotes {
		newAggSig.AggregateSig(&v.Signature)
	}
	mv.aggregateSignature = newAggSig
	return shouldRemove
}

func newMempoolVote(outOf uint32, voteData *primitives.VoteData) *mempoolVote {
	return &mempoolVote{
		participationBitfield: make([]uint8, (outOf+7)/8),
		aggregateSignature:    bls.NewAggregateSignature(),
		individualVotes:       make([]*primitives.SingleValidatorVote, 0, outOf),
		voteData:              voteData,
	}
}

type VoteMempool struct {
	poolLock sync.RWMutex
	pool     map[chainhash.Hash]*mempoolVote
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

func pickPercent(vs []primitives.SingleValidatorVote, pct float32) []primitives.SingleValidatorVote {
	num := int(pct * float32(len(vs)))
	shuffledVotes := shuffleVotes(vs)
	return shuffledVotes[:num]
}

func (m *VoteMempool) GetVotesNotInBloom(bloom *bloom.BloomFilter) []primitives.SingleValidatorVote {
	votes := make([]primitives.SingleValidatorVote, 0)
	for _, vs := range m.pool {
		for _, v := range vs.individualVotes {
			vh := v.Hash()
			if bloom.Has(vh) {
				continue
			}

			votes = append(votes, *v)
		}
	}
	return votes
}

func (m *VoteMempool) Add(vote *primitives.SingleValidatorVote, outOf uint32) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	voteHash := vote.Data.Hash()

	if vs, found := m.pool[voteHash]; found {
		vs.add(vote)
	} else {
		m.pool[voteHash] = newMempoolVote(outOf, &vote.Data)
		m.pool[voteHash].add(vote)
	}
}

func (m *VoteMempool) Get(slot uint64, p *params.ChainParams) []primitives.MultiValidatorVote {
	votes := make([]primitives.MultiValidatorVote, 0)
	for i := range m.pool {
		if m.pool[i].voteData.Slot < slot-p.MinAttestationInclusionDelay {
			vote := primitives.MultiValidatorVote{
				Data:                  *m.pool[i].voteData,
				Signature:             *m.pool[i].aggregateSignature,
				ParticipationBitfield: append([]uint8(nil), m.pool[i].participationBitfield...),
			}
			votes = append(votes, vote)
		}
	}

	return votes
}

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
		}
	}

}

// NewMempool creates a new mempool.
func NewMempool() *VoteMempool {
	return &VoteMempool{
		pool: make(map[chainhash.Hash]*mempoolVote),
	}
}
