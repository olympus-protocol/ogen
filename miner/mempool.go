package miner

import (
	"sync"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type Mempool struct {
	poolLock sync.RWMutex
	pool     map[chainhash.Hash]*primitives.MultiValidatorVote
}

func (m *Mempool) add(vote *primitives.SingleValidatorVote, outOf uint32) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	voteHash := vote.Data.Hash()

	if v, found := m.pool[voteHash]; found {
		if v.ParticipationBitfield[vote.Offset/8]&(1<<uint(vote.Offset%8)) > 0 {
			// we already have this vote
			return
		}
		v.Signature.AggregateSig(&vote.Signature)
		v.ParticipationBitfield[vote.Offset/8] |= (1 << uint(vote.Offset%8))
	} else {
		participationBitfield := make([]uint8, (outOf+7)/8)
		participationBitfield[vote.Offset/8] |= (1 << uint(vote.Offset%8))
		m.pool[voteHash] = &primitives.MultiValidatorVote{
			Data:                  vote.Data,
			Signature:             vote.Signature,
			ParticipationBitfield: participationBitfield,
		}
	}
}

func (m *Mempool) get(slot uint64, p *params.ChainParams) []primitives.MultiValidatorVote {
	votes := make([]primitives.MultiValidatorVote, 0)
	for i := range m.pool {
		if m.pool[i].Data.Slot < slot-p.MinAttestationInclusionDelay {
			votes = append(votes, *m.pool[i])
		}
	}

	return votes
}

func (m *Mempool) remove(b *primitives.Block) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	for _, v := range b.Votes {
		voteHash := v.Data.Hash()

		if vote, found := m.pool[voteHash]; found {
		inner:
			for i := range vote.ParticipationBitfield {
				if vote.ParticipationBitfield[i]&v.ParticipationBitfield[i] != 0 {
					delete(m.pool, voteHash)
					break inner
				}
			}
		}
	}

}

// NewMempool creates a new mempool.
func NewMempool() *Mempool {
	return &Mempool{
		pool: make(map[chainhash.Hash]*primitives.MultiValidatorVote),
	}
}
