package primitives_test

import (
	"github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestState_Copy(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(100, 100)
	balances := map[[20]byte]uint64{}
	nonces := map[[20]byte]uint64{}
	f.Fuzz(&balances)
	f.Fuzz(&nonces)

	key := [20]byte{1, 2, 3, 200}

	cs := primitives.CoinsState{
		Balances: balances,
		Nonces:   nonces,
	}

	cs.Balances[key] = 10
	cs.Nonces[key] = 10

	f.NilChance(0).NumElements(5, 5)

	keyRep := [20]byte{1, 2, 200}
	keyCom := [32]byte{1, 2, 3}

	replace := map[[20]byte]chainhash.Hash{}
	community := map[chainhash.Hash]primitives.CommunityVoteData{}
	f.Fuzz(&replace)
	f.Fuzz(&community)

	gs := primitives.Governance{
		ReplaceVotes:   replace,
		CommunityVotes: community,
	}

	gs.ReplaceVotes[keyRep] = [32]byte{200}
	gs.CommunityVotes[keyCom] = primitives.CommunityVoteData{
		ReplacementCandidates: [][20]byte{
			{1, 2, 3},
		},
	}

	var valreg []*primitives.Validator
	f.NilChance(0).NumElements(100, 100)
	f.Fuzz(&valreg)

	valreg[0].LastActiveEpoch = 0

	voteInfoSlice := []*primitives.AcceptedVoteInfo{
		{
			Data: &primitives.VoteData{
				Slot:            0,
				FromEpoch:       0,
				FromHash:        [32]byte{},
				ToEpoch:         0,
				ToHash:          [32]byte{},
				BeaconBlockHash: [32]byte{},
				Nonce:           0,
			},
			ParticipationBitfield: bitfield.NewBitlist(4 * 8),
			Proposer:              10,
			InclusionDelay:        5,
		},
	}

	latestBlocksSlice := [][32]byte{
		{1, 2, 3},
	}
	currManagersSlice := [][20]byte{
		{1, 2, 3},
	}
	s := primitives.State{

		CoinsState:                    cs,
		Governance:                    gs,
		ValidatorRegistry:             valreg,
		ManagerReplacement:            bitfield.NewBitlist(5 * 8),
		LatestBlockHashes:             latestBlocksSlice,
		CurrentManagers:               currManagersSlice,
		RANDAO:                        [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		NextRANDAO:                    [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		PreviousJustifiedEpochHash:    [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		JustifiedEpochHash:            [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		CurrentEpochVotes:             voteInfoSlice,
		PreviousEpochVotes:            voteInfoSlice,
		ProposerQueue:                 []uint64{0, 1, 2, 3, 4, 5},
		PreviousEpochVoteAssignments:  []uint64{0, 1, 2, 3, 4, 5},
		CurrentEpochVoteAssignments:   []uint64{0, 1, 2, 3, 4, 5},
		NextProposerQueue:             []uint64{0, 1, 2, 3, 4, 5},
		LatestValidatorRegistryChange: 5,
		Slot:                          5,
		EpochIndex:                    5,
		JustificationBitfield:         5,
		FinalizedEpoch:                5,
		JustifiedEpoch:                5,
		PreviousJustifiedEpoch:        5,
		VoteEpoch:                     5,
		VoteEpochStartSlot:            5,
		VotingState:                   5,
		LastPaidSlot:                  5,
	}

	s2 := s.Copy()

	s.CoinsState.Nonces[key] = 11
	assert.Equal(t, s2.CoinsState.Nonces[key], uint64(10))

	s.CoinsState.Balances[key] = 11
	assert.Equal(t, s2.CoinsState.Balances[key], uint64(10))

	s.Governance.ReplaceVotes[keyRep] = [32]byte{201}
	s2RepVote := s2.Governance.ReplaceVotes[keyRep]
	expBytes := [32]byte{200}
	assert.Equal(t, s2RepVote[:], expBytes[:])

	s.Governance.CommunityVotes[keyCom] = primitives.CommunityVoteData{
		ReplacementCandidates: [][20]byte{
			{1, 2, 5},
		},
	}
	assert.Equal(t, s2.Governance.CommunityVotes[keyCom].ReplacementCandidates, [][20]byte{{1, 2, 3}})

	s.ValidatorRegistry[0].LastActiveEpoch = 1

	assert.Equal(t, s2.ValidatorRegistry[0].LastActiveEpoch, uint64(0))

	s.LatestValidatorRegistryChange = 6
	assert.Equal(t, s2.LatestValidatorRegistryChange, uint64(5))

	s.Slot = 6
	assert.Equal(t, s2.Slot, uint64(5))

	s.EpochIndex = 6
	assert.Equal(t, s2.EpochIndex, uint64(5))

	s.JustificationBitfield = 6
	assert.Equal(t, s2.JustificationBitfield, uint64(5))

	s.FinalizedEpoch = 6
	assert.Equal(t, s2.FinalizedEpoch, uint64(5))

	s.JustifiedEpoch = 6
	assert.Equal(t, s2.JustifiedEpoch, uint64(5))

	s.PreviousJustifiedEpoch = 6
	assert.Equal(t, s2.PreviousJustifiedEpoch, uint64(5))

	s.VoteEpoch = 6
	assert.Equal(t, s2.VoteEpoch, uint64(5))

	s.VoteEpochStartSlot = 6
	assert.Equal(t, s2.VoteEpochStartSlot, uint64(5))

	s.VotingState = 6
	assert.Equal(t, s2.VotingState, uint64(5))

	s.LastPaidSlot = 6
	assert.Equal(t, s2.LastPaidSlot, uint64(5))

	s.ProposerQueue[5] = 6
	assert.Equal(t, s2.ProposerQueue[5], uint64(5))

	s.PreviousEpochVoteAssignments[5] = 6
	assert.Equal(t, s2.PreviousEpochVoteAssignments[5], uint64(5))

	s.CurrentEpochVoteAssignments[5] = 6
	assert.Equal(t, s2.CurrentEpochVoteAssignments[5], uint64(5))

	s.NextProposerQueue[5] = 6
	assert.Equal(t, s2.NextProposerQueue[5], uint64(5))

	s.RANDAO[4] = 6
	assert.Equal(t, s2.RANDAO[4], uint8(5))

	s.NextRANDAO[4] = 6
	assert.Equal(t, s2.NextRANDAO[4], uint8(5))

	s.PreviousJustifiedEpochHash[4] = 6
	assert.Equal(t, s2.PreviousJustifiedEpochHash[4], uint8(5))

	s.JustifiedEpochHash[4] = 6
	assert.Equal(t, s2.JustifiedEpochHash[4], uint8(5))

	s.CurrentEpochVotes[0].Proposer = 6
	assert.Equal(t, s2.CurrentEpochVotes[0].Proposer, uint64(10))

	s.PreviousEpochVotes[0].Proposer = 6
	assert.Equal(t, s2.PreviousEpochVotes[0].Proposer, uint64(10))

	s.LatestBlockHashes[0][0] = 2
	assert.Equal(t, s2.LatestBlockHashes[0][0], uint8(1))

	s.CurrentManagers[0][0] = 2
	assert.Equal(t, s2.CurrentManagers[0][0], uint8(1))

	assert.Equal(t, s.ManagerReplacement, s2.ManagerReplacement)

	s.ManagerReplacement[0] = 1
	assert.Equal(t, s2.ManagerReplacement[0], uint8(0))

	assert.Equal(t, s.ManagerReplacement.Len(), s2.ManagerReplacement.Len())

	assert.Equal(t, len(s.ManagerReplacement), len(s2.ManagerReplacement))

}

func TestState_FromSerializable(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(100, 100)
	balances := map[[20]byte]uint64{}
	nonces := map[[20]byte]uint64{}
	f.Fuzz(&balances)
	f.Fuzz(&nonces)

	cs := primitives.CoinsState{
		Balances: balances,
		Nonces:   nonces,
	}
	scs := cs.ToSerializable()

	f.NilChance(0).NumElements(5, 5)

	replace := map[[20]byte]chainhash.Hash{}
	community := map[chainhash.Hash]primitives.CommunityVoteData{}
	f.Fuzz(&replace)
	f.Fuzz(&community)

	gs := primitives.Governance{
		ReplaceVotes:   replace,
		CommunityVotes: community,
	}
	sgs := gs.ToSerializable()

	var randao, nextrandao, justifiedepoch, previousjustified, governance chainhash.Hash
	f.Fuzz(&randao)
	f.Fuzz(&nextrandao)
	f.Fuzz(&justifiedepoch)
	f.Fuzz(&previousjustified)
	f.Fuzz(&governance)

	var latestRegistry, slot, epoch, justifbit, finalepoch, justified uint64
	var previousjustepoch, voteepoch, votestartslot, votestate, lastpayedslot uint64

	f.Fuzz(&latestRegistry)
	f.Fuzz(&slot)
	f.Fuzz(&epoch)
	f.Fuzz(&justifbit)
	f.Fuzz(&finalepoch)
	f.Fuzz(&justifiedepoch)
	f.Fuzz(&previousjustepoch)
	f.Fuzz(&voteepoch)
	f.Fuzz(&votestartslot)
	f.Fuzz(&votestate)
	f.Fuzz(&lastpayedslot)
	f.Fuzz(&justified)

	var valreg []*primitives.Validator
	f.NumElements(100, 100)
	f.Fuzz(&valreg)

	var currManagers [][20]byte
	f.NumElements(5, 5)
	f.Fuzz(&currManagers)

	var currEpochVotes, prevEpochVotes []*primitives.AcceptedVoteInfo
	f.NumElements(20, 20)
	f.Fuzz(&currEpochVotes)
	f.Fuzz(&prevEpochVotes)

	var latesBlockHashes [][32]byte
	f.NumElements(64, 64)
	f.Fuzz(&latesBlockHashes)

	var proposerQueue, prevEpochVoteAssign, currEpochVoteAssign, nextPropQueue []uint64
	f.NumElements(100, 100)
	f.Fuzz(&proposerQueue)
	f.Fuzz(&prevEpochVoteAssign)
	f.Fuzz(&currEpochVoteAssign)
	f.Fuzz(&nextPropQueue)

	sst := &primitives.SerializableState{

		CoinsState: &scs,
		Governance: &sgs,

		ValidatorRegistry:             valreg,
		LatestValidatorRegistryChange: latestRegistry,
		RANDAO:                        randao,
		NextRANDAO:                    nextrandao,
		Slot:                          slot,
		EpochIndex:                    epoch,
		ProposerQueue:                 proposerQueue,
		PreviousEpochVoteAssignments:  prevEpochVoteAssign,
		CurrentEpochVoteAssignments:   currEpochVoteAssign,
		NextProposerQueue:             nextPropQueue,
		JustificationBitfield:         justifbit,
		FinalizedEpoch:                finalepoch,
		LatestBlockHashes:             latesBlockHashes,
		JustifiedEpoch:                justified,
		JustifiedEpochHash:            justifiedepoch,
		CurrentEpochVotes:             currEpochVotes,
		PreviousJustifiedEpoch:        previousjustepoch,
		PreviousJustifiedEpochHash:    previousjustified,
		PreviousEpochVotes:            prevEpochVotes,
		CurrentManagers:               currManagers,
		ManagerReplacement:            bitfield.NewBitlist(5 * 8),
		VoteEpoch:                     voteepoch,
		VoteEpochStartSlot:            votestartslot,
		VotingState:                   votestate,
		LastPaidSlot:                  lastpayedslot,
	}

	st := new(primitives.State)
	st.FromSerializable(sst)

	assert.Equal(t, len(st.ValidatorRegistry), len(sst.ValidatorRegistry))

	assert.Equal(t, st.LatestValidatorRegistryChange, sst.LatestValidatorRegistryChange)

	assert.Equal(t, st.RANDAO[:], sst.RANDAO[:])

	assert.Equal(t, st.NextRANDAO[:], sst.NextRANDAO[:])

	assert.Equal(t, st.Slot, sst.Slot)

	assert.Equal(t, st.EpochIndex, sst.EpochIndex)

	assert.Equal(t, len(st.ProposerQueue), len(sst.ProposerQueue))

	assert.Equal(t, len(st.PreviousEpochVoteAssignments), len(sst.PreviousEpochVoteAssignments))

	assert.Equal(t, len(st.CurrentEpochVoteAssignments), len(sst.CurrentEpochVoteAssignments))

	assert.Equal(t, len(st.NextProposerQueue), len(sst.NextProposerQueue))

	assert.Equal(t, st.JustificationBitfield, sst.JustificationBitfield)

	assert.Equal(t, st.FinalizedEpoch, sst.FinalizedEpoch)

	assert.Equal(t, len(st.LatestBlockHashes), len(sst.LatestBlockHashes))

	assert.Equal(t, st.JustifiedEpoch, sst.JustifiedEpoch)

	assert.Equal(t, st.JustifiedEpochHash[:], sst.JustifiedEpochHash[:])

	assert.Equal(t, len(st.CurrentEpochVotes), len(sst.CurrentEpochVotes))

	assert.Equal(t, st.PreviousJustifiedEpoch, sst.PreviousJustifiedEpoch)

	assert.Equal(t, st.PreviousJustifiedEpochHash[:], sst.PreviousJustifiedEpochHash[:])

	assert.Equal(t, len(st.PreviousEpochVotes), len(sst.PreviousEpochVotes))

	assert.Equal(t, len(st.CurrentManagers), len(sst.CurrentManagers))

	assert.Equal(t, st.ManagerReplacement[:], sst.ManagerReplacement[:])

	assert.Equal(t, st.VoteEpoch, sst.VoteEpoch)

	assert.Equal(t, st.VoteEpochStartSlot, sst.VoteEpochStartSlot)

	assert.Equal(t, st.VotingState, sst.VotingState)

	assert.Equal(t, st.LastPaidSlot, sst.LastPaidSlot)
}

func TestState_ToSerializable(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(100, 100)
	balances := map[[20]byte]uint64{}
	nonces := map[[20]byte]uint64{}
	f.Fuzz(&balances)
	f.Fuzz(&nonces)

	cs := primitives.CoinsState{
		Balances: balances,
		Nonces:   nonces,
	}

	f.NilChance(0).NumElements(5, 5)

	replace := map[[20]byte]chainhash.Hash{}
	community := map[chainhash.Hash]primitives.CommunityVoteData{}
	f.Fuzz(&replace)
	f.Fuzz(&community)

	gs := primitives.Governance{
		ReplaceVotes:   replace,
		CommunityVotes: community,
	}
	var randao, nextrandao, justifiedepoch, previousjustified, governance chainhash.Hash
	f.Fuzz(&randao)
	f.Fuzz(&nextrandao)
	f.Fuzz(&justifiedepoch)
	f.Fuzz(&previousjustified)
	f.Fuzz(&governance)

	var latestRegistry, slot, epoch, justifbit, finalepoch, justified uint64
	var previousjustepoch, voteepoch, votestartslot, votestate, lastpayedslot uint64

	f.Fuzz(&latestRegistry)
	f.Fuzz(&slot)
	f.Fuzz(&epoch)
	f.Fuzz(&justifbit)
	f.Fuzz(&finalepoch)
	f.Fuzz(&justifiedepoch)
	f.Fuzz(&previousjustepoch)
	f.Fuzz(&voteepoch)
	f.Fuzz(&votestartslot)
	f.Fuzz(&votestate)
	f.Fuzz(&lastpayedslot)
	f.Fuzz(&justified)

	var valreg []*primitives.Validator
	f.NumElements(100, 100)
	f.Fuzz(&valreg)

	var currManagers [][20]byte
	f.NumElements(5, 5)
	f.Fuzz(&currManagers)

	var currEpochVotes, prevEpochVotes []*primitives.AcceptedVoteInfo
	f.NumElements(20, 20)
	f.Fuzz(&currEpochVotes)
	f.Fuzz(&prevEpochVotes)

	var latesBlockHashes [][32]byte
	f.NumElements(64, 64)
	f.Fuzz(&latesBlockHashes)

	var proposerQueue, prevEpochVoteAssign, currEpochVoteAssign, nextPropQueue []uint64
	f.NumElements(100, 100)
	f.Fuzz(&proposerQueue)
	f.Fuzz(&prevEpochVoteAssign)
	f.Fuzz(&currEpochVoteAssign)
	f.Fuzz(&nextPropQueue)

	st := primitives.State{

		CoinsState: cs,
		Governance: gs,

		ValidatorRegistry:             valreg,
		LatestValidatorRegistryChange: latestRegistry,
		RANDAO:                        randao,
		NextRANDAO:                    nextrandao,
		Slot:                          slot,
		EpochIndex:                    epoch,
		ProposerQueue:                 proposerQueue,
		PreviousEpochVoteAssignments:  prevEpochVoteAssign,
		CurrentEpochVoteAssignments:   currEpochVoteAssign,
		NextProposerQueue:             nextPropQueue,
		JustificationBitfield:         justifbit,
		FinalizedEpoch:                finalepoch,
		LatestBlockHashes:             latesBlockHashes,
		JustifiedEpoch:                justified,
		JustifiedEpochHash:            justifiedepoch,
		CurrentEpochVotes:             currEpochVotes,
		PreviousJustifiedEpoch:        previousjustepoch,
		PreviousJustifiedEpochHash:    previousjustified,
		PreviousEpochVotes:            prevEpochVotes,
		CurrentManagers:               currManagers,
		ManagerReplacement:            bitfield.NewBitlist(5 * 8),
		VoteEpoch:                     voteepoch,
		VoteEpochStartSlot:            votestartslot,
		VotingState:                   votestate,
		LastPaidSlot:                  lastpayedslot,
	}

	sst := st.ToSerializable()

	assert.Equal(t, len(st.ValidatorRegistry), len(sst.ValidatorRegistry))

	assert.Equal(t, st.LatestValidatorRegistryChange, sst.LatestValidatorRegistryChange)

	assert.Equal(t, st.RANDAO[:], sst.RANDAO[:])

	assert.Equal(t, st.NextRANDAO[:], sst.NextRANDAO[:])

	assert.Equal(t, st.Slot, sst.Slot)

	assert.Equal(t, st.EpochIndex, sst.EpochIndex)

	assert.Equal(t, len(st.ProposerQueue), len(sst.ProposerQueue))

	assert.Equal(t, len(st.PreviousEpochVoteAssignments), len(sst.PreviousEpochVoteAssignments))

	assert.Equal(t, len(st.CurrentEpochVoteAssignments), len(sst.CurrentEpochVoteAssignments))

	assert.Equal(t, len(st.NextProposerQueue), len(sst.NextProposerQueue))

	assert.Equal(t, st.JustificationBitfield, sst.JustificationBitfield)

	assert.Equal(t, st.FinalizedEpoch, sst.FinalizedEpoch)

	assert.Equal(t, len(st.LatestBlockHashes), len(sst.LatestBlockHashes))

	assert.Equal(t, st.JustifiedEpoch, sst.JustifiedEpoch)

	assert.Equal(t, st.JustifiedEpochHash[:], sst.JustifiedEpochHash[:])

	assert.Equal(t, len(st.CurrentEpochVotes), len(sst.CurrentEpochVotes))

	assert.Equal(t, st.PreviousJustifiedEpoch, sst.PreviousJustifiedEpoch)

	assert.Equal(t, st.PreviousJustifiedEpochHash[:], sst.PreviousJustifiedEpochHash[:])

	assert.Equal(t, len(st.PreviousEpochVotes), len(sst.PreviousEpochVotes))

	assert.Equal(t, len(st.CurrentManagers), len(sst.CurrentManagers))

	assert.Equal(t, st.ManagerReplacement[:], sst.ManagerReplacement[:])

	assert.Equal(t, st.VoteEpoch, sst.VoteEpoch)

	assert.Equal(t, st.VoteEpochStartSlot, sst.VoteEpochStartSlot)

	assert.Equal(t, st.VotingState, sst.VotingState)

	assert.Equal(t, st.LastPaidSlot, sst.LastPaidSlot)

}
