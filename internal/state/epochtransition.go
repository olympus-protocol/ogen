package state

import (
	"bytes"
	"errors"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/bls/multisig"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"math/big"
)

// GetEffectiveBalance gets the balance of a validator.
func (s *state) GetEffectiveBalance(index uint64) uint64 {
	netParams := config.GlobalParams.NetParams

	b := s.ValidatorRegistry[index].Balance
	if b >= netParams.DepositAmount {
		return netParams.DepositAmount
	}

	return b
}

func (s *state) getActiveBalance() uint64 {
	balance := uint64(0)

	for _, v := range s.ValidatorRegistry {
		if !v.IsActive() {
			continue
		}

		balance += v.Balance
	}

	return balance
}

// ActivateValidator activates a validator in the state at a certain index.
func (s *state) ActivateValidator(index uint64) error {
	validator := s.ValidatorRegistry[index]
	if validator.Status != primitives.StatusStarting {
		return errors.New("validator is not pending activation")
	}

	validator.Status = primitives.StatusActive
	return nil
}

// InitiateValidatorExit moves a validator from active to pending exit.
func (s *state) InitiateValidatorExit(index uint64) error {
	validator := s.ValidatorRegistry[index]
	if validator.Status != primitives.StatusActive {
		return errors.New("validator is not active")
	}

	validator.Status = primitives.StatusActivePendingExit
	return nil
}

// ExitValidator handles state changes when a validator exits.
func (s *state) ExitValidator(index uint64, status uint64) error {
	netParams := config.GlobalParams.NetParams

	validator := s.ValidatorRegistry[index]
	prevStatus := validator.Status

	if prevStatus == primitives.StatusExitedWithPenalty {
		return nil
	}

	validator.Status = status

	if status == primitives.StatusExitedWithPenalty {
		slotIndex := (s.Slot + netParams.EpochLength - 1) % netParams.EpochLength

		proposerIndex := s.ProposerQueue[slotIndex]

		whistleblowerReward := s.GetEffectiveBalance(proposerIndex) / netParams.WhistleblowerRewardQuotient

		s.ValidatorRegistry[proposerIndex].Balance += whistleblowerReward
		s.ValidatorRegistry[index].Balance -= whistleblowerReward

		return nil
	}
	s.CoinsState.Balances[validator.PayeeAddress] += validator.Balance
	validator.Balance = 0

	return nil
}

// UpdateValidatorStatus moves a validator to a specific status.
func (s *state) UpdateValidatorStatus(index uint64, status uint64) error {
	if status == primitives.StatusActive {
		err := s.ActivateValidator(index)
		return err
	} else if status == primitives.StatusActivePendingExit {
		err := s.InitiateValidatorExit(index)
		return err
	} else if status == primitives.StatusExitedWithPenalty || status == primitives.StatusExitedWithoutPenalty {
		err := s.ExitValidator(index, status)
		return err
	}
	return nil
}

func (s *state) updateValidatorRegistry() error {
	netParams := config.GlobalParams.NetParams

	totalBalance := s.getActiveBalance()

	// 1/2 balance churn goes to starting validators and 1/2 goes to exiting validators
	maxBalanceChurn := totalBalance / (netParams.MaxBalanceChurnQuotient * 2)

	balanceChurn := uint64(0)
	for idx, validator := range s.ValidatorRegistry {
		index := uint64(idx)

		// start validators if needed
		if validator.Status == primitives.StatusStarting && validator.Balance == netParams.DepositAmount*netParams.UnitsPerCoin && validator.FirstActiveEpoch <= s.EpochIndex {
			balanceChurn += s.GetEffectiveBalance(index)

			if balanceChurn > maxBalanceChurn {
				break
			}

			err := s.UpdateValidatorStatus(index, primitives.StatusActive)
			if err != nil {
				return err
			}
		}
	}

	balanceChurn = 0
	for idx, validator := range s.ValidatorRegistry {
		index := uint64(idx)

		if validator.Status == primitives.StatusActivePendingExit && validator.LastActiveEpoch <= s.EpochIndex {
			balanceChurn += s.GetEffectiveBalance(index)

			if balanceChurn > maxBalanceChurn {
				break
			}

			err := s.UpdateValidatorStatus(index, primitives.StatusExitedWithoutPenalty)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetRecentBlockHash gets block hashes from the LatestBlockHashes array.
func (s *state) GetRecentBlockHash(slotToGet uint64) chainhash.Hash {
	netParams := config.GlobalParams.NetParams

	if s.Slot-slotToGet >= netParams.LatestBlockRootsLength {
		return chainhash.Hash{}
	}

	return s.LatestBlockHashes[slotToGet%netParams.LatestBlockRootsLength]
}

// GetTotalBalances gets the total balances of the state.
func (s *state) GetTotalBalances() uint64 {
	total := uint64(0)
	for _, v := range s.ValidatorRegistry {
		total += v.Balance
	}

	total += s.CoinsState.GetTotal()

	return total
}

// NextVoteEpoch increments the voting epoch, resets votes,
// and updates the state.
func (s *state) NextVoteEpoch(newState uint64) {
	s.VoteEpoch++
	s.VoteEpochStartSlot = s.Slot
	// TODO reinitiate the governance state.

	s.VotingState = newState
}

// CheckForVoteTransitions tallies up votes and checks for any governance
// state transitions.
func (s *state) CheckForVoteTransitions() {
	netParams := config.GlobalParams.NetParams

	switch s.VotingState {
	case GovernanceStateActive:
		// if it's active, we should check if we've accumulated enough votes
		// to start a community vote
		totalBalance := s.GetTotalBalances()
		votingBalance := uint64(0)
		for acc := range s.Governance.ReplaceVotes {
			bal := s.CoinsState.Balances[acc]
			votingBalance += bal
		}

		if votingBalance*netParams.CommunityOverrideQuotient >= totalBalance {
			s.NextVoteEpoch(GovernanceStateVoting)
			for i := range s.CurrentManagers {
				s.ManagerReplacement.Set(uint(i))
			}
		}
	case GovernanceStateVoting:
		if s.VoteEpochStartSlot+netParams.VotingPeriodSlots <= s.Slot {
			// tally votes and choose next managers
			managerVotes := make(map[chainhash.Hash]uint64)

			for acc, rpv := range s.Governance.ReplaceVotes {
				bal := s.CoinsState.Balances[acc]
				if _, ok := managerVotes[rpv]; ok {
					managerVotes[rpv] += bal
				} else {
					managerVotes[rpv] = bal
				}
			}

			bestBalance := uint64(0)
			bestManagers := s.CurrentManagers
			for i, v := range managerVotes {
				if v > bestBalance {
					voteData := s.Governance.CommunityVotes[i]

					newManagers := make([][20]byte, len(s.CurrentManagers))
					copy(newManagers, s.CurrentManagers)

					for i := range newManagers {
						if s.ManagerReplacement.Get(uint(i)) {
							copy(newManagers[i][:], voteData.ReplacementCandidates[i][:])
						}
					}

					bestManagers = newManagers
				}
			}

			s.CurrentManagers = bestManagers
			s.NextVoteEpoch(GovernanceStateActive)
		}
	}

	// process payouts if needed
	epochsPerMonth := 30 * 24 * 60 * 60 / netParams.SlotDuration / netParams.EpochLength
	if s.LastPaidSlot/netParams.EpochLength+epochsPerMonth <= s.Slot {
		// 10% to 5/5 multisig
		// 10% to each

		totalBlockReward := netParams.BaseRewardPerBlock * 60 * 60 * 24 * 30 / netParams.SlotDuration
		perGroup := totalBlockReward / 10

		multipub := multisig.PublicKeyHashesToMultisigHash(s.CurrentManagers, 5)
		s.CoinsState.Balances[multipub] += perGroup
		if len(s.CurrentManagers) != len(netParams.GovernancePercentages) {
			return
		}

		for group, address := range s.CurrentManagers {
			percent := netParams.GovernancePercentages[group]
			s.CoinsState.Balances[address] += perGroup * uint64(percent) / 100
		}

		s.LastPaidSlot = s.Slot
	}
}

// ProcessEpochTransition runs an epoch transition on the state.
func (s *state) ProcessEpochTransition() ([]*primitives.EpochReceipt, error) {
	netParams := config.GlobalParams.NetParams

	s.CheckForVoteTransitions()

	totalBalance := s.getActiveBalance()

	// These are voters who voted for a target of the previous epoch.
	previousEpochVoters := newVoterGroup()
	previousEpochVotersMatchingTargetHash := newVoterGroup()
	previousEpochVotersMatchingBeaconBlock := newVoterGroup()

	currentEpochVotersMatchingTarget := newVoterGroup()

	previousEpochBoundaryHash := chainhash.Hash{}
	if s.Slot >= 2*netParams.EpochLength {
		previousEpochBoundaryHash = s.GetRecentBlockHash(s.Slot - 2*netParams.EpochLength - 1)
	}

	epochBoundaryHash := chainhash.Hash{}
	if s.Slot >= netParams.EpochLength {
		epochBoundaryHash = s.GetRecentBlockHash(s.Slot - netParams.EpochLength - 1)
	}

	// previousEpochVotersMap maps validator to their assigned vote

	previousEpochVotersMap := make(map[uint64]*primitives.AcceptedVoteInfo)

	for _, v := range s.PreviousEpochVotes {
		validatorIndices, err := s.GetVoteCommittee(v.Data.Slot)
		if err != nil {
			return nil, err
		}
		previousEpochVoters.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield, validatorIndices)
		actualBlockHash := s.GetRecentBlockHash(v.Data.Slot - 1)
		if bytes.Equal(actualBlockHash[:], v.Data.BeaconBlockHash[:]) {
			previousEpochVotersMatchingBeaconBlock.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield, validatorIndices)
		}
		if bytes.Equal(previousEpochBoundaryHash[:], v.Data.ToHash[:]) {
			previousEpochVotersMatchingTargetHash.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield, validatorIndices)
		}
		for _, validatorIdx := range validatorIndices {
			previousEpochVotersMap[validatorIdx] = v
		}
	}

	for _, v := range s.CurrentEpochVotes {
		validators, err := s.GetVoteCommittee(v.Data.Slot)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(epochBoundaryHash[:], v.Data.FromHash[:]) {
			currentEpochVotersMatchingTarget.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield, validators)
		}
	}

	s.PreviousJustifiedEpoch = s.JustifiedEpoch
	s.PreviousJustifiedEpochHash = s.JustifiedEpochHash
	s.JustificationBitfield <<= 1

	// >2/3 voted with target of the previous epoch
	if 3*previousEpochVotersMatchingTargetHash.totalBalance >= 2*totalBalance {
		s.JustificationBitfield |= 1 << 1 // mark
		s.JustifiedEpoch = s.EpochIndex - 1
		s.JustifiedEpochHash = s.GetRecentBlockHash(s.JustifiedEpoch * netParams.EpochLength)
	}

	if 3*currentEpochVotersMatchingTarget.totalBalance >= 2*totalBalance {
		s.JustificationBitfield |= 1 << 0
		s.JustifiedEpoch = s.EpochIndex
		s.JustifiedEpochHash = s.GetRecentBlockHash(s.JustifiedEpoch * netParams.EpochLength)
	}

	if (s.JustificationBitfield>>1)%4 == 3 && s.PreviousJustifiedEpoch == s.EpochIndex-2 { // 110 <- old previous justified would be
		s.FinalizedEpoch = s.PreviousJustifiedEpoch
	}

	if ((s.JustificationBitfield>>0)%8 == 7 && s.JustifiedEpoch == s.EpochIndex-1) || // 111
		((s.JustificationBitfield>>0)%4 == 3 && s.JustifiedEpoch == s.EpochIndex) {
		s.FinalizedEpoch = s.PreviousJustifiedEpoch
		s.JustifiedEpoch = s.EpochIndex
	}

	// TODO move to params
	const numRewards = 5

	baseReward := func(index uint64) uint64 {
		return s.GetEffectiveBalance(index) * netParams.UnitsPerCoin * netParams.BaseRewardPerBlock * netParams.EpochLength / totalBalance / numRewards
	}

	receipts := make([]*primitives.EpochReceipt, 0)

	rewardValidator := func(index uint64, reward uint64, why uint64) {
		s.ValidatorRegistry[index].Balance += reward
		receipts = append(receipts, &primitives.EpochReceipt{
			Validator: index,
			Amount:    int64(reward),
			Type:      why,
		})
	}

	penalizeValidator := func(index uint64, penalty uint64, why uint64) {
		if s.ValidatorRegistry[index].FirstActiveEpoch+5 >= s.EpochIndex {
			return
		}
		s.ValidatorRegistry[index].Balance -= penalty
		receipts = append(receipts, &primitives.EpochReceipt{
			Validator: index,
			Amount:    -int64(penalty),
			Type:      why,
		})
	}

	if s.Slot >= 2*netParams.EpochLength {

		for index, validator := range s.ValidatorRegistry {
			idx := uint64(index)
			if !validator.IsActive() {
				continue
			}

			// votes matching source rewarded
			if previousEpochVoters.contains(idx) {
				reward := baseReward(idx)
				rewardValidator(idx, reward, primitives.RewardMatchedFromEpoch)
			} else {
				penalty := baseReward(idx)
				penalizeValidator(idx, penalty, primitives.PenaltyMissingFromEpoch)
			}

			// votes matching target rewarded
			if previousEpochVotersMatchingTargetHash.contains(idx) {
				reward := baseReward(idx)
				rewardValidator(idx, reward, primitives.RewardMatchedToEpoch)
			} else {
				penalty := baseReward(idx)
				penalizeValidator(idx, penalty, primitives.PenaltyMissingToEpoch)
			}

			// votes matching beacon block rewarded
			if previousEpochVotersMatchingBeaconBlock.contains(idx) {
				reward := baseReward(idx)
				rewardValidator(idx, reward, primitives.RewardMatchedBeaconBlock)
			} else {
				penalty := baseReward(idx)
				penalizeValidator(idx, penalty, primitives.PenaltyMissingBeaconBlock)
			}
		}

		// inclusion rewards
		proposerRewardInclusion := make(map[uint64]uint64)
		proposerRewardDistance := make(map[uint64]uint64)
		for voter := range previousEpochVoters.voters {
			vote := previousEpochVotersMap[voter]

			proposerIndex := vote.Proposer

			reward := baseReward(proposerIndex)
			if _, ok := proposerRewardInclusion[proposerIndex]; !ok {
				proposerRewardInclusion[proposerIndex] = 0
			}
			proposerRewardInclusion[proposerIndex] += reward

			inclusionDistance := vote.InclusionDelay

			reward = baseReward(proposerIndex) * netParams.MinAttestationInclusionDelay / inclusionDistance
			if _, ok := proposerRewardDistance[proposerIndex]; !ok {
				proposerRewardDistance[proposerIndex] = 0
			}
			proposerRewardDistance[proposerIndex] += reward
		}

		for validator, amount := range proposerRewardInclusion {
			rewardValidator(validator, amount, primitives.RewardIncludedVote)
		}
		for validator, amount := range proposerRewardDistance {
			rewardValidator(validator, amount, primitives.RewardInclusionDistance)
		}

		// Penalize all validators after 4 epochs and punish validators who did not vote more severely.
		finalityDelay := s.EpochIndex - s.FinalizedEpoch
		if finalityDelay > 4 {
			for index := range s.ValidatorRegistry {
				idx := uint64(index)
				w := s.ValidatorRegistry[idx]
				if !w.IsActive() {
					continue
				}

				penalty := baseReward(idx) * numRewards

				penalizeValidator(idx, penalty, primitives.PenaltyInactivityLeak)

				if !previousEpochVotersMatchingTargetHash.contains(idx) {
					penalty := s.GetEffectiveBalance(idx) * finalityDelay / netParams.InactivityPenaltyQuotient
					penalizeValidator(idx, penalty, primitives.PenaltyInactivityLeakNoVote)
				}
			}
		}
	}

	for index, validator := range s.ValidatorRegistry {
		if validator.IsActive() && validator.Balance < netParams.EjectionBalance*netParams.UnitsPerCoin {
			err := s.UpdateValidatorStatus(uint64(index), primitives.StatusExitedWithoutPenalty)
			if err != nil {
				return nil, err
			}
		}
	}

	s.EpochIndex = s.Slot / netParams.EpochLength

	if s.FinalizedEpoch > s.LatestValidatorRegistryChange {
		err := s.updateValidatorRegistry()
		if err != nil {
			return nil, err
		}

		s.LatestValidatorRegistryChange = s.EpochIndex
	}

	s.ProposerQueue = s.NextProposerQueue
	activeValidators := s.GetValidatorIndicesActiveAt(s.EpochIndex + 1)
	s.NextProposerQueue = DetermineNextProposers(s.RANDAO, activeValidators)

	s.PreviousEpochVoteAssignments = s.CurrentEpochVoteAssignments
	s.CurrentEpochVoteAssignments = Shuffle(s.RANDAO, activeValidators)

	copy(s.RANDAO[:], s.NextRANDAO[:])

	s.PreviousEpochVotes = s.CurrentEpochVotes
	s.CurrentEpochVotes = make([]*primitives.AcceptedVoteInfo, 0)

	return receipts, nil
}

func generateRandNumber(from chainhash.Hash, max uint32) uint64 {
	randaoBig := new(big.Int)
	randaoBig.SetBytes(from[:])

	numValidator := big.NewInt(int64(max + 1))

	return randaoBig.Mod(randaoBig, numValidator).Uint64()
}

// Shuffle shuffles validator using a RANDAO.
func Shuffle(randao chainhash.Hash, vals []uint64) []uint64 {
	nextProposers := make([]uint64, len(vals))
	copy(nextProposers, vals)

	for i := uint64(0); i < uint64(len(nextProposers)-1); i++ {
		j := i + generateRandNumber(randao, uint32(len(nextProposers))-uint32(i)-1)
		nextProposers[i], nextProposers[j] = nextProposers[j], nextProposers[i]
		randao = chainhash.HashH(randao[:])
	}

	return nextProposers
}

// DetermineNextProposers gets the next shuffling.
func DetermineNextProposers(randao chainhash.Hash, activeValidators []uint64) []uint64 {
	netParams := config.GlobalParams.NetParams

	validatorsChosen := make(map[uint64]struct{})
	nextProposers := make([]uint64, netParams.EpochLength)

	for i := range nextProposers {
		found := true
		var val uint64

		for found {
			val = generateRandNumber(randao, uint32(len(activeValidators)-1))
			randao = chainhash.HashH(randao[:])
			_, found = validatorsChosen[val]
		}

		validatorsChosen[val] = struct{}{}
		nextProposers[i] = val
		randao = chainhash.HashH(randao[:])
	}

	return nextProposers
}
