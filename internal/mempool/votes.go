package mempool

import (
	"bytes"
	"context"
	"errors"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
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
	"sync"
	"time"
)

// VoteMempool is the interface of the voteMempool
type VoteMempool interface {
	AddValidate(vote *primitives.MultiValidatorVote, s state.State) error
	Add(vote *primitives.MultiValidatorVote)
	Get(slot uint64, s state.State, proposerIndex uint64) ([]*primitives.MultiValidatorVote, error)
	Remove(b *primitives.Block)
	Notify(notifee VoteSlashingNotifee)
}

type voteMempoolItem struct {
	vote *primitives.MultiValidatorVote
}

// voteMempool is a mempool that keeps track of votes.
type voteMempool struct {
	poolLock sync.Mutex
	pool     map[chainhash.Hash]*primitives.MultiValidatorVote

	netParams *params.ChainParams
	log       logger.Logger
	ctx       context.Context

	chain chain.Blockchain
	host  hostnode.HostNode

	notifees     []VoteSlashingNotifee
	notifeesLock sync.Mutex

	lastActionManager actionmanager.LastActionManager
}

var _ VoteMempool = &voteMempool{}

// AddValidate validates, then adds the vote to the mempool.
func (m *voteMempool) AddValidate(vote *primitives.MultiValidatorVote, s state.State) error {
	if err := s.IsVoteValid(vote); err != nil {
		return err
	}
	m.Add(vote)
	return nil
}

// Add adds a vote to the mempool.
func (m *voteMempool) Add(vote *primitives.MultiValidatorVote) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()
	voteData := vote.Data
	voteHash := voteData.Hash()

	firstSlotAllowedToInclude := vote.Data.Slot + m.netParams.MinAttestationInclusionDelay

	currentState, err := m.chain.State().TipStateAtSlot(firstSlotAllowedToInclude)
	if err != nil {
		m.log.Error(err)
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

		if currentState.GetSlot() >= v.Data.LastSlotValid(m.netParams) {
			delete(m.pool, voteHash)
			continue
		}

		var votingValidators = make(map[uint64]struct{})
		var common []uint64

		vote1Committee, err := currentState.GetVoteCommittee(v.Data.Slot)
		if err != nil {
			m.log.Error(err)
		}

		vote2Committee, err := currentState.GetVoteCommittee(vote.Data.Slot)
		if err != nil {
			m.log.Error(err)
		}

		for i, idx := range vote1Committee {
			if !v.ParticipationBitfield.Get(uint(i)) {
				continue
			}
			votingValidators[idx] = struct{}{}
		}

		for i, idx := range vote2Committee {
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
		m.log.Debugf("received vote with same vote data aggregating %d votes...", len(vote.ParticipationBitfield.BitIndices()))
		if !bytes.Equal(v.Sig[:], vote.Sig[:]) {

			// Check if votes overlaps voters
			voteCommittee, err := currentState.GetVoteCommittee(v.Data.Slot)
			if err != nil {
				m.log.Error(err)
			}

			var common []uint64

			for i, idx := range voteCommittee {
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
			}

			sig1, err := bls.SignatureFromBytes(v.Sig[:])
			if err != nil {
				m.log.Error(err)
			}

			sig2, err := bls.SignatureFromBytes(vote.Sig[:])
			if err != nil {
				m.log.Error(err)
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
		m.log.Debugf("adding vote to the mempool with %d votes", len(vote.ParticipationBitfield.BitIndices()))
		m.pool[voteHash] = vote
	}
}

// Get gets a vote from the mempool.
func (m *voteMempool) Get(slot uint64, s state.State, proposerIndex uint64) ([]*primitives.MultiValidatorVote, error) {
	m.poolLock.Lock()
	defer m.poolLock.Unlock()

	votes := make([]*primitives.MultiValidatorVote, 0)

	for _, vote := range m.pool {

		if slot >= vote.Data.FirstSlotValid(m.netParams) && slot <= vote.Data.LastSlotValid(m.netParams) {
			err := s.ProcessVote(vote, proposerIndex)
			if err != nil {
				m.log.Error(err)
				m.poolLock.Lock()
				voteHash := vote.Data.Hash()
				delete(m.pool, voteHash)
				m.poolLock.Unlock()
				continue
			}
			if uint64(len(votes)) < m.netParams.MaxVotesPerBlock {
				votes = append(votes, vote)
			}
		}

	}

	return votes, nil
}

// Remove removes mempool items that are no longer relevant.
func (m *voteMempool) Remove(b *primitives.Block) {
	netParams := config.GlobalParams.NetParams
	m.poolLock.Lock()
	defer m.poolLock.Unlock()

	// Check for votes on the block and remove them
	for _, v := range b.Votes {
		voteHash := v.Data.Hash()

		// If the vote is on pool and included on the block, remove it.
		poolVote, ok := m.pool[voteHash]
		if ok {
			m.log.Debugf("removing vote from mempool block vote contains %d votes and vote on mempool contains %d votes", v.ParticipationBitfield.Len(), poolVote.ParticipationBitfield.Len())
			delete(m.pool, voteHash)
			m.log.Debugf("votes on pool %d", len(m.pool))
		}

		if b.Header.Slot >= v.Data.LastSlotValid(netParams) {
			delete(m.pool, voteHash)
		}
	}

}

func (m *voteMempool) getCurrentSlot() uint64 {
	slot := time.Now().Sub(m.chain.GenesisTime()) / (time.Duration(m.netParams.SlotDuration) * time.Second)
	if slot < 0 {
		return 0
	}
	return uint64(slot)
}

func (m *voteMempool) handleVote(id peer.ID, msg p2p.Message) error {

	if id == m.host.GetHost().ID() {
		return nil
	}

	data, ok := msg.(*p2p.MsgVote)
	if !ok {
		return errors.New("wrong message on vote topic")
	}

	vote := data.Data

	firstSlotAllowedToInclude := vote.Data.Slot + m.netParams.MinAttestationInclusionDelay
	tip := m.chain.State().Tip()

	if tip.Slot+m.netParams.EpochLength*2 < firstSlotAllowedToInclude {
		return nil
	}

	view, err := m.chain.State().GetSubView(tip.Hash)
	if err != nil {
		m.log.Warnf("could not get block view representing current tip: %s", err)
		return err
	}

	currentState, _, err := m.chain.State().GetStateForHashAtSlot(tip.Hash, firstSlotAllowedToInclude, &view)
	if err != nil {
		m.log.Warnf("error updating chain to attestation inclusion slot: %s", err)
		return err
	}
	m.log.Debugf("received vote from %s with %d votes", id, len(data.Data.ParticipationBitfield.BitIndices()))
	err = m.AddValidate(data.Data, currentState)
	if err != nil {

		return err
	}

	return nil
}

// Notify registers a notifee to be notified when illegal votes occur.
func (m *voteMempool) Notify(notifee VoteSlashingNotifee) {
	m.notifeesLock.Lock()
	defer m.notifeesLock.Unlock()
	m.notifees = append(m.notifees, notifee)
}

// NewVoteMempool creates a new mempool.
func NewVoteMempool(ch chain.Blockchain, hostnode hostnode.HostNode, manager actionmanager.LastActionManager) (VoteMempool, error) {
	ctx := config.GlobalParams.Context
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	vm := &voteMempool{
		pool:              make(map[chainhash.Hash]*primitives.MultiValidatorVote),
		netParams:         netParams,
		log:               log,
		ctx:               ctx,
		chain:             ch,
		notifees:          make([]VoteSlashingNotifee, 0),
		host:              hostnode,
		lastActionManager: manager,
	}

	if err := vm.host.RegisterTopicHandler(p2p.MsgVoteCmd, vm.handleVote); err != nil {
		return nil, err
	}

	return vm, nil
}

// VoteSlashingNotifee is notified when an illegal vote occurs.
type VoteSlashingNotifee interface {
	NotifyIllegalVotes(slashing *primitives.VoteSlashing)
}
