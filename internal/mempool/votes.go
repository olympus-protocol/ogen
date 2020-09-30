package mempool

import (
	"bytes"
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"sort"
	"sync"
	"time"
)

// VoteMempool is the interface of the voteMempool
type VoteMempool interface {
	AddValidate(vote *primitives.MultiValidatorVote, state state.State) error
	Add(vote *primitives.MultiValidatorVote)
	Get(slot uint64, s state.State, p *params.ChainParams, proposerIndex uint64) ([]*primitives.MultiValidatorVote, error)
	Remove(b *primitives.Block)
	Notify(notifee VoteSlashingNotifee)
}

// voteMempool is a mempool that keeps track of votes.
type voteMempool struct {
	poolLock sync.Mutex
	pool     map[chainhash.Hash]*primitives.MultiValidatorVote

	// chainindex 0 is highest prioritized, 1 is less, etc
	poolOrder []chainhash.Hash

	params     *params.ChainParams
	log        logger.Logger
	ctx        context.Context
	blockchain chain.Blockchain
	hostNode   hostnode.HostNode
	voteTopic  *pubsub.Topic

	notifees     []VoteSlashingNotifee
	notifeesLock sync.Mutex

	lastActionManager actionmanager.LastActionManager
}

var _ VoteMempool = &voteMempool{}

// AddValidate validates, then adds the vote to the mempool.
func (m *voteMempool) AddValidate(vote *primitives.MultiValidatorVote, state state.State) error {
	if err := state.IsVoteValid(vote, m.params); err != nil {
		return err
	}

	m.Add(vote)
	return nil
}

// sortMempool sorts the poolOrder so that the highest priority transactions come first and assumes you hold the poolLock.
func (m *voteMempool) sortMempool() {
	sort.Slice(m.poolOrder, func(i, j int) bool {

		// return if i is higher priority than j

		iHash := m.poolOrder[i]
		jHash := m.poolOrder[j]
		iData := m.pool[iHash].Data
		jData := m.pool[jHash].Data

		// first sort by slot
		if iData.Slot < jData.Slot {
			return true
		} else if iData.Slot > jData.Slot {
			return false
		}

		// arbitrary
		return true
	})
}

// Add adds a vote to the mempool.
func (m *voteMempool) Add(vote *primitives.MultiValidatorVote) {
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
	for h, v := range m.pool {

		// If the vote data hash matches, it means is voting for same block.
		if voteHash.IsEqual(&h) {
			continue
		}

		if currentState.GetSlot() >= v.Data.LastSlotValid(m.params) {
			delete(m.pool, voteHash)
			m.removeFromOrder(voteHash)
			continue
		}

		var votingValidators = make(map[uint64]struct{})
		var common []uint64
		vote1Comitte, err := currentState.GetVoteCommittee(v.Data.Slot, m.params)
		if err != nil {
			m.log.Error(err)
			return
		}
		vote2Comitte, err := currentState.GetVoteCommittee(vote.Data.Slot, m.params)
		if err != nil {
			m.log.Error(err)
			return
		}
		for i, idx := range vote1Comitte {
			if !v.ParticipationBitfield.Get(uint(i)) {
				continue
			}
			votingValidators[idx] = struct{}{}
		}
		for i, idx := range vote2Comitte {
			if !vote.ParticipationBitfield.Get(uint(i)) {
				continue
			}
			_, exist := votingValidators[idx]
			if exist {
				common = append(common, idx)
			}
		}

		// Check if the new vote with different hash overlaps a previous marked validator vote.
		if len(common) != 0 {
			// Check if the vote matches the double vote and surround vote conditions
			// If there is an intersection check if is a double vote
			if v.Data.IsSurroundVote(voteData) || v.Data.IsDoubleVote(voteData) {
				// If is a double or surround vote announce it and slash.
				if v.Data.IsSurroundVote(voteData) {
					m.log.Warnf("found surround vote for multivalidator, reporting...")
				}
				if v.Data.IsDoubleVote(voteData) {
					m.log.Warnf("found double vote for multivalidator, reporting...")
				}
				for _, n := range m.notifees {
					n.NotifyIllegalVotes(&primitives.VoteSlashing{
						Vote1: vote,
						Vote2: v,
					})
				}
				return
			}
		}
	}

	// Check if vote is already on pool.
	// If a vote with same vote data is found we should check the signatures.
	// If the signatures are the same it means is a duplicated vote for network (probably a relayed vote).
	// If the signatures don't match, we should aggregate both signatures and merge the bitlists.
	// IMPORTANT: 	We should never allow a vote that conflicts a previous vote to be added to the pool.
	// 				That should be checked against all votes on pool comparing bitlists.
	v, ok := m.pool[voteHash]

	if ok {
		if !bytes.Equal(v.Sig[:], vote.Sig[:]) {
			// Check if votes overlaps voters

			votesComitte, err := currentState.GetVoteCommittee(v.Data.Slot, m.params)
			if err != nil {
				m.log.Error(err)
				return
			}

			var common []uint64

			for i, idx := range votesComitte {
				if v.ParticipationBitfield.Get(uint(i)) && vote.ParticipationBitfield.Get(uint(i)) {
					common = append(common, idx)
				}
			}

			if len(common) != 0 {
				// If the vote overlaps, that means a validator submitted the same vote multiple times.
				for _, n := range m.notifees {
					n.NotifyIllegalVotes(&primitives.VoteSlashing{
						Vote1: vote,
						Vote2: v,
					})
				}
				return
			}

			newBitfield, err := v.ParticipationBitfield.Merge(vote.ParticipationBitfield)
			if err != nil {
				m.log.Error(err)
				return
			}
			sig1, err := bls.SignatureFromBytes(v.Sig[:])
			if err != nil {
				m.log.Error(err)
				return
			}
			sig2, err := bls.SignatureFromBytes(vote.Sig[:])
			if err != nil {
				m.log.Error(err)
				return
			}
			newVoteSig := bls.AggregateSignatures([]*bls.Signature{sig1, sig2})

			var voteSig [96]byte
			copy(voteSig[:], newVoteSig.Marshal())

			newVote := &primitives.MultiValidatorVote{
				Data:                  v.Data,
				ParticipationBitfield: newBitfield,
				Sig:                   voteSig,
			}

			m.pool[voteHash] = newVote
		}
	} else {

		m.pool[voteHash] = vote
		m.poolOrder = append(m.poolOrder, voteHash)
	}

	m.sortMempool()
}

// Get gets a vote from the mempool.
func (m *voteMempool) Get(slot uint64, s state.State, p *params.ChainParams, proposerIndex uint64) ([]*primitives.MultiValidatorVote, error) {
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
func (m *voteMempool) removeFromOrder(h chainhash.Hash) {
	newOrder := make([]chainhash.Hash, 0, len(m.poolOrder))

	for _, vh := range m.poolOrder {
		if !vh.IsEqual(&h) {
			newOrder = append(newOrder, vh)
		}
	}

	m.poolOrder = newOrder
}

// Remove removes mempool items that are no longer relevant.
func (m *voteMempool) Remove(b *primitives.Block) {
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

}

func (m *voteMempool) getCurrentSlot() uint64 {
	slot := time.Now().Sub(m.blockchain.GenesisTime()) / (time.Duration(m.params.SlotDuration) * time.Second)
	if slot < 0 {
		return 0
	}
	return uint64(slot)
}

func (m *voteMempool) handleSubscription(sub *pubsub.Subscription, id peer.ID) {
	for {
		msg, err := sub.Next(m.ctx)
		if err != nil {
			if err != m.ctx.Err() {
				m.log.Warnf("error getting next message in votes topic: %s", err)
				continue
			}
		}

		if msg.GetFrom() == id {
			continue
		}

		buf := bytes.NewBuffer(msg.Data)

		voteRead, err := p2p.ReadMessage(buf, m.hostNode.GetNetMagic())
		if err != nil {
			m.log.Warnf("unable to decode message: %s", err)
			continue
		}

		voteMsg, ok := voteRead.(*p2p.MsgVote)
		if !ok {
			m.log.Warnf("peer sent wrong message on vote subscription")
			continue
		}

		vote := voteMsg.Data

		m.log.Debugf("received votes from peer %s", msg.GetFrom().String())

		firstSlotAllowedToInclude := vote.Data.Slot + m.params.MinAttestationInclusionDelay
		tip := m.blockchain.State().Tip()

		if tip.Slot+m.params.EpochLength*2 < firstSlotAllowedToInclude {
			continue
		}

		view, err := m.blockchain.State().GetSubView(tip.Hash)
		if err != nil {
			m.log.Warnf("could not get block view representing current tip: %s", err)
			continue
		}

		currentState, _, err := m.blockchain.State().GetStateForHashAtSlot(tip.Hash, firstSlotAllowedToInclude, &view, m.params)
		if err != nil {
			m.log.Warnf("error updating chain to attestation inclusion slot: %s", err)
			continue
		}

		err = m.AddValidate(vote, currentState)
		if err != nil {
			m.log.Debugf("error adding vote to mempool: %s", err)
			continue
		}
	}
}

// Notify registers a notifee to be notified when illegal votes occur.
func (m *voteMempool) Notify(notifee VoteSlashingNotifee) {
	m.notifeesLock.Lock()
	defer m.notifeesLock.Unlock()
	m.notifees = append(m.notifees, notifee)
}

// NewVoteMempool creates a new mempool.
func NewVoteMempool(ctx context.Context, log logger.Logger, p *params.ChainParams, ch chain.Blockchain, hostnode hostnode.HostNode, manager actionmanager.LastActionManager) (VoteMempool, error) {
	voteTopic, err := hostnode.Topic(p2p.MsgVoteCmd)
	if err != nil {
		return nil, err
	}

	voteSub, err := voteTopic.Subscribe()
	if err != nil {
		return nil, err
	}

	vm := &voteMempool{
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
