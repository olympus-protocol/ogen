package mempool

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/peers/conflict"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/logger"
	"sort"
	"sync"
)

type mempoolVote struct {
	individualVotes         []*primitives.SingleValidatorVote
	participationBitfield   bitfield.Bitlist
	participatingValidators map[uint64]struct{}
	voteData                *primitives.VoteData
}

func (mv *mempoolVote) getVoteByOffset(offset uint64) (*primitives.SingleValidatorVote, bool) {
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

func (mv *mempoolVote) add(vote *primitives.SingleValidatorVote, voter uint64) {
	if mv.participationBitfield.Get(uint(vote.Offset)) {
		return
	}
	mv.individualVotes = append(mv.individualVotes, vote)
	mv.participationBitfield.Set(uint(vote.Offset))
	mv.participatingValidators[voter] = struct{}{}
}

func (mv *mempoolVote) remove(participationBitfield bitfield.Bitlist) (shouldRemove bool) {
	shouldRemove = true
	newVotes := make([]*primitives.SingleValidatorVote, 0, len(mv.individualVotes))
	for _, v := range mv.individualVotes {
		if uint64(len(participationBitfield)*8) >= v.Offset {
			return shouldRemove
		}
		if !participationBitfield.Get(uint(v.Offset)) {
			newVotes = append(newVotes, v)
			shouldRemove = false
		}
	}

	mv.individualVotes = newVotes
	mv.participationBitfield = bitfield.NewBitlist(participationBitfield.Len())
	for i, p := range participationBitfield {
		mv.participationBitfield[i] = p
	}
	return shouldRemove
}

func newMempoolVote(outOf uint64, voteData *primitives.VoteData) *mempoolVote {
	return &mempoolVote{
		participationBitfield:   bitfield.NewBitlist(outOf + 7),
		individualVotes:         make([]*primitives.SingleValidatorVote, 0, outOf),
		voteData:                voteData,
		participatingValidators: make(map[uint64]struct{}),
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
	hostNode   *peers.HostNode
	voteTopic  *pubsub.Topic

	notifees     []VoteSlashingNotifee
	notifeesLock sync.Mutex

	lastActionManager *conflict.LastActionManager
}

// AddValidate validates, then adds the vote to the mempool.
func (m *VoteMempool) AddValidate(vote *primitives.SingleValidatorVote, state *primitives.State) error {
	if err := state.IsVoteValid(vote.AsMulti(), m.params); err != nil {
		return err
	}

	m.Add(vote)
	return nil
}

// sortMempool sorts the poolOrder so that the highest priority transactions come first and assumes you hold the poolLock.
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
	currentState, err := m.blockchain.State().TipStateAtSlot(firstSlotAllowedToInclude)
	if err != nil {
		m.log.Error(err)
		return
	}
	committee, err := currentState.GetVoteCommittee(vote.Data.Slot, m.params)
	if err != nil {
		m.log.Error(err)
		return
	}

	if vote.Offset >= uint64(len(committee)) {
		return
	}

	voter := committee[vote.Offset]

	m.lastActionManager.RegisterAction(currentState.ValidatorRegistry[voter].PubKey, vote.Data.Nonce)

	// slashing check... check if this vote interferes with any votes in the mempool
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

				voteMulti := vote.AsMulti()
				conflictingMulti := conflicting.AsMulti()
				for _, n := range m.notifees {
					n.NotifyIllegalVotes(&primitives.VoteSlashing{
						Vote1: voteMulti,
						Vote2: conflictingMulti,
					})
				}
				return
			}
		}
	}

	if vs, found := m.pool[voteHash]; found {
		vs.add(vote, voter)
	} else {
		m.pool[voteHash] = newMempoolVote(vote.OutOf, vote.Data)
		m.poolOrder = append(m.poolOrder, voteHash)
		m.pool[voteHash].add(vote, voter)
	}

	m.sortMempool()
}

// Get gets a vote from the mempool.
func (m *VoteMempool) Get(slot uint64, s *primitives.State, p *params.ChainParams, proposerIndex uint64) ([]*primitives.MultiValidatorVote, error) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	votes := make([]*primitives.MultiValidatorVote, 0)
	for _, h := range m.poolOrder {
		v := m.pool[h]
		if slot >= v.voteData.FirstSlotValid(p) && slot <= v.voteData.LastSlotValid(p) {
			sigs := make([]*bls.Signature, 0)
			for _, v := range m.pool[h].individualVotes {
				sig, err := v.Signature()
				if err != nil {
					return nil, err
				}
				sigs = append(sigs, sig)
			}
			sig := bls.AggregateSignatures(sigs)
			var sigb [96]byte
			copy(sigb[:], sig.Marshal())
			bl := bitfield.NewBitlist(v.participationBitfield.Len())
			for i, p := range v.participationBitfield {
				bl[i] = p
			}
			vote := &primitives.MultiValidatorVote{
				Data:                  m.pool[h].voteData,
				Sig:                   sigb,
				ParticipationBitfield: bl,
			}
			if err := s.ProcessVote(vote, p, proposerIndex); err != nil {
				return nil, err
			}
			if uint64(len(votes)) < p.MaxVotesPerBlock {
				votes = append(votes, vote)
			}
		}
	}

	return votes, nil
}

// Assumes you already hold the poolLock mutex.
func (m *VoteMempool) removeFromOrder(h chainhash.Hash) {
	newOrder := make([]chainhash.Hash, 0, len(m.poolOrder))

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

		if b.Header.Slot >= v.Data.LastSlotValid(m.params) {
			delete(m.pool, voteHash)
			m.removeFromOrder(voteHash)
		}
	}
}

func (m *VoteMempool) handleSubscription(sub *pubsub.Subscription, id peer.ID) {
	for {
		msg, err := sub.Next(m.ctx)
		if err != nil {
			m.log.Warnf("error getting next message in votes topic: %s", err)
			return
		}

		if msg.GetFrom() == id {
			continue
		}

		votes := new(primitives.Votes)

		if err := votes.Unmarshal(msg.Data); err != nil {
			m.log.Warnf("peer sent invalid vote: %s", err)
			err = m.hostNode.BanScorePeer(msg.GetFrom(), peers.BanLimit)
			if err == nil {
				m.log.Warnf("peer %s was banned", msg.GetFrom().String())
			}
			continue
		}

		m.log.Debugf("received votes msg with %d votes", len(votes.Votes))
		var wg sync.WaitGroup
		wg.Add(len(votes.Votes))
		for _, v := range votes.Votes {
			go func(vote *primitives.SingleValidatorVote, wg *sync.WaitGroup) {
				defer wg.Done()
				firstSlotAllowedToInclude := vote.Data.Slot + m.params.MinAttestationInclusionDelay
				tip := m.blockchain.State().Tip()

				if tip.Slot+m.params.EpochLength*2 < firstSlotAllowedToInclude {
					return
				}

				view, err := m.blockchain.State().GetSubView(tip.Hash)
				if err != nil {
					m.log.Warnf("could not get block view representing current tip: %s", err)
					return
				}
				currentState, _, err := m.blockchain.State().GetStateForHashAtSlot(tip.Hash, firstSlotAllowedToInclude, &view, m.params)
				if err != nil {
					m.log.Warnf("error updating chain to attestation inclusion slot: %s", err)
					return
				}

				err = m.AddValidate(vote, currentState)
				if err != nil {
					m.log.Debugf("error adding transaction to mempool (might not be synced): %s", err)
				}
			}(v, &wg)
		}
		wg.Wait()
	}
}

// Notify registers a notifee to be notified when illegal votes occur.
func (m *VoteMempool) Notify(notifee VoteSlashingNotifee) {
	m.notifeesLock.Lock()
	defer m.notifeesLock.Unlock()
	m.notifees = append(m.notifees, notifee)
}

// NewVoteMempool creates a new mempool.
func NewVoteMempool(ctx context.Context, log *logger.Logger, p *params.ChainParams, ch *chain.Blockchain, hostnode *peers.HostNode, manager *conflict.LastActionManager) (*VoteMempool, error) {
	voteTopic, err := hostnode.Topic("votes")
	if err != nil {
		return nil, err
	}

	voteSub, err := voteTopic.Subscribe()
	if err != nil {
		return nil, err
	}

	_, err = voteTopic.Relay()
	if err != nil {
		return nil, err
	}

	vm := &VoteMempool{
		pool:              make(map[chainhash.Hash]*mempoolVote),
		params:            p,
		log:               log,
		ctx:               ctx,
		blockchain:        ch,
		voteTopic:         voteTopic,
		notifees:          make([]VoteSlashingNotifee, 0),
		hostNode:          hostnode,
		lastActionManager: manager,
	}

	go vm.handleSubscription(voteSub, hostnode.GetHost().ID())

	return vm, nil
}

// VoteSlashingNotifee is notified when an illegal vote occurs.
type VoteSlashingNotifee interface {
	NotifyIllegalVotes(slashing *primitives.VoteSlashing)
}
