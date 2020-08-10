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

// VoteMempool is a mempool that keeps track of votes.
type VoteMempool struct {
	poolLock sync.Mutex
	pool     map[chainhash.Hash]*primitives.MultiValidatorVote

	// index 0 is highest prioritized, 1 is less, etc
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
func (m *VoteMempool) AddValidate(vote *primitives.MultiValidatorVote, state *primitives.State) error {
	if err := state.IsVoteValid(vote, m.params); err != nil {
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
		iData := m.pool[iHash].Data
		iNumVotes := len(m.pool[iHash].ParticipationBitfield)
		jData := m.pool[jHash].Data
		jNumVotes := len(m.pool[iHash].ParticipationBitfield)

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
func (m *VoteMempool) Add(vote *primitives.MultiValidatorVote) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	voteData := vote.Data
	voteHash := voteData.Hash()


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

	if len(vote.ParticipationBitfield) >= len(committee) {
		return
	}

	// Register voting action for validators included on the vote
	for _, c := range committee {
		if vote.ParticipationBitfield.Get(uint(c)) {
			m.lastActionManager.RegisterAction(currentState.ValidatorRegistry[c].PubKey, vote.Data.Nonce)
		}
	}

	// Slashing check. Check if the vote interferes with any vote on the pool.
	for h, v := range m.pool {

		// If the vote data hash matches, it means is voting for same block.
		if voteHash.IsEqual(&h) {
			continue
		}

		if v.Data.IsSurroundVote(voteData) {
			m.log.Warnf("found surround vote for multivalidator in vote %s ...", vote.Data.String())
		}

		if v.Data.IsDoubleVote(voteData) {
			m.log.Warnf("found double vote for multivalidator in vote %s ...", vote.Data.String())
		}

		for _, n := range m.notifees {
			n.NotifyIllegalVotes(&primitives.VoteSlashing{
				Vote1: vote,
				Vote2: v,
			})
		}
		return

	}

	// Check if vote is already on mempool.
	// If a vote with same vote data is found, aggregate signatures and add it to the mempool
	// Check if vote is already on mempool.
	v, ok := m.pool[voteHash]
	if ok {

		newVote := &primitives.MultiValidatorVote{
			Data:                  v.Data,
			Sig:                   [96]byte{},
			ParticipationBitfield: bitfield.NewBitlist(uint64(len(committee) * 8)),
		}

		for _, c := range committee {
			if v.ParticipationBitfield.Get(uint(c)) || vote.ParticipationBitfield.Get(uint(c)) {
				newVote.ParticipationBitfield.Set(uint(c))
			}
		}
		sig := v.Sig
		newSig := vote.Sig
		newVoteSig, err := bls.AggregateSignaturesBytes([][96]byte{sig, newSig})
		if err != nil {
			m.log.Error(err)
			return
		}
		var voteSig [96]byte
		copy(voteSig[:], newVoteSig.Marshal())
		newVote.Sig = voteSig
		m.pool[voteHash] = newVote

	} else {
		m.pool[voteHash] = vote
		m.poolOrder = append(m.poolOrder, voteHash)
	}

	m.sortMempool()
}

// Get gets a vote from the mempool.
func (m *VoteMempool) Get(slot uint64, s *primitives.State, p *params.ChainParams, proposerIndex uint64) ([]*primitives.MultiValidatorVote, error) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	votes := make([]*primitives.MultiValidatorVote, 0)
	for _, h := range m.poolOrder {
		vote := m.pool[h]
		if slot >= vote.Data.FirstSlotValid(p) && slot <= vote.Data.LastSlotValid(p) {
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

	// Check for votes on the block and remove them
	for _, v := range b.Votes {
		voteHash := v.Data.Hash()

		// If the vote is on pool and included on the block, remove it.
		_, ok := m.pool[voteHash]
		if ok {
			delete(m.pool, voteHash)
			m.removeFromOrder(voteHash)
		}
	}

	// Check all votes against the block slot to remove expired votes
	for _, voteHash := range m.poolOrder {
		v, ok := m.pool[voteHash]
		if ok {
			if b.Header.Slot >= v.Data.LastSlotValid(m.params) {
				delete(m.pool, voteHash)
				m.removeFromOrder(voteHash)
			}
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

		vote := new(primitives.MultiValidatorVote)

		if err := vote.Unmarshal(msg.Data); err != nil {
			m.log.Warnf("peer sent invalid vote: %s", err)
			err = m.hostNode.BanScorePeer(msg.GetFrom(), peers.BanLimit)
			if err == nil {
				m.log.Warnf("peer %s was banned", msg.GetFrom().String())
			}
			continue
		}

		m.log.Debugf("received votes from peer %s", id.String())

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
		pool:              make(map[chainhash.Hash]*primitives.MultiValidatorVote),
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
