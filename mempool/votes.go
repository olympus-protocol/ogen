package mempool

import (
	"bytes"
	"context"
	"math/rand"
	"sort"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
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
	// TODO: validate vote

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

	if vs, found := m.pool[voteHash]; found {
		vs.add(vote)
	} else {
		m.pool[voteHash] = newMempoolVote(vote.OutOf, &vote.Data)
		m.poolOrder = append(m.poolOrder, voteHash)
		m.pool[voteHash].add(vote)
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
			vote := primitives.MultiValidatorVote{
				Data:                  *m.pool[i].voteData,
				Signature:             *m.pool[i].aggregateSignature,
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

func (m *VoteMempool) handleSubscription(topic *pubsub.Subscription) {
	for {
		msg, err := topic.Next(m.ctx)
		if err != nil {
			m.log.Warnf("error getting next message in votes topic: %s", err)
			return
		}

		txBuf := bytes.NewReader(msg.Data)
		tx := new(primitives.SingleValidatorVote)

		if err := tx.Decode(txBuf); err != nil {
			// TODO: ban peer
			m.log.Warnf("peer sent invalid vote: %s", err)
			continue
		}

		currentState := m.blockchain.State().TipState()

		err = m.AddValidate(tx, currentState)
		if err != nil {
			m.log.Warnf("error adding transaction to mempool: %s", err)
		}
	}
}

// NewVoteMempool creates a new mempool.
func NewVoteMempool(ctx context.Context, log *logger.Logger, p *params.ChainParams, ch *chain.Blockchain, hostnode *peers.HostNode) (*VoteMempool, error) {
	vm := &VoteMempool{
		pool:       make(map[chainhash.Hash]*mempoolVote),
		params:     p,
		log:        log,
		ctx:        ctx,
		blockchain: ch,
	}
	voteTopic, err := hostnode.Topic("votes")
	if err != nil {
		return nil, err
	}

	voteSub, err := voteTopic.Subscribe()
	if err != nil {
		return nil, err
	}

	go vm.handleSubscription(voteSub)

	return vm, nil
}
