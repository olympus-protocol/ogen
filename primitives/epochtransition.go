package primitives

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/logger"
)

// GetEffectiveBalance gets the balance of a validator.
func (s *State) GetEffectiveBalance(index uint32, p *params.ChainParams) uint64 {
	b := s.ValidatorRegistry[index].Balance
	if b >= p.DepositAmount {
		return p.DepositAmount
	}

	return b
}

func (s *State) getActiveBalance(p *params.ChainParams) uint64 {
	balance := uint64(0)

	for _, v := range s.ValidatorRegistry {
		if !v.IsActive() {
			continue
		}

		balance += v.Balance
	}

	return balance
}

func intSqrt(n uint64) uint64 {
	x := n
	y := (x + 1) / 2
	for y < x {
		x = y
		y = (x + n/x) / 2
	}
	return x
}

type voterGroup struct {
	voters       map[uint32]struct{}
	totalBalance uint64
}

func (vg *voterGroup) add(id uint32, bal uint64) {
	if _, found := vg.voters[id]; found {
		return
	}

	vg.voters[id] = struct{}{}
	vg.totalBalance += bal
}

func (vg *voterGroup) addFromBitfield(registry []Validator, bitfield []uint8, validatorIndices []uint32) {
	for idx, validatorIdx := range validatorIndices {
		b := idx / 8
		j := idx % 8
		if bitfield[b]&(1<<j) > 0 {
			vg.add(validatorIdx, registry[validatorIdx].Balance)
		}
	}
}

func (vg *voterGroup) contains(id uint32) bool {
	_, found := vg.voters[id]
	return found
}

func newVoterGroup() voterGroup {
	return voterGroup{
		voters: make(map[uint32]struct{}),
	}
}

// ActivateValidator activates a validator in the state at a certain index.
func (s *State) ActivateValidator(index uint32) error {
	validator := &s.ValidatorRegistry[index]
	if validator.Status != StatusStarting {
		return errors.New("validator is not pending activation")
	}

	validator.Status = StatusActive
	return nil
}

// InitiateValidatorExit moves a validator from active to pending exit.
func (s *State) InitiateValidatorExit(index uint32) error {
	validator := &s.ValidatorRegistry[index]
	if validator.Status != StatusActive {
		return errors.New("validator is not active")
	}

	validator.Status = StatusActivePendingExit
	return nil
}

// ExitValidator handles state changes when a validator exits.
func (s *State) ExitValidator(index uint32, status uint8, p *params.ChainParams) error {
	validator := &s.ValidatorRegistry[index]
	prevStatus := validator.Status

	if prevStatus == StatusExitedWithPenalty {
		return nil
	}

	validator.Status = status

	if status == StatusExitedWithPenalty {
		slotIndex := (s.Slot + p.EpochLength - 1) % p.EpochLength

		proposerIndex := s.ProposerQueue[slotIndex]

		whistleblowerReward := s.GetEffectiveBalance(proposerIndex, p) / p.WhistleblowerRewardQuotient

		s.ValidatorRegistry[proposerIndex].Balance += whistleblowerReward
		s.ValidatorRegistry[index].Balance -= whistleblowerReward

		return nil
	}

	s.CoinsState.Balances[validator.PayeeAddress] += validator.Balance
	validator.Balance = 0

	return nil
}

// UpdateValidatorStatus moves a validator to a specific status.
func (s *State) UpdateValidatorStatus(index uint32, status uint8, p *params.ChainParams) error {
	if status == StatusActive {
		err := s.ActivateValidator(index)
		return err
	} else if status == StatusActivePendingExit {
		err := s.InitiateValidatorExit(index)
		return err
	} else if status == StatusExitedWithPenalty || status == StatusExitedWithoutPenalty {
		err := s.ExitValidator(index, status, p)
		return err
	}
	return nil
}

func (s *State) updateValidatorRegistry(p *params.ChainParams) error {
	totalBalance := s.getActiveBalance(p)

	// 1/2 balance churn goes to starting validators and 1/2 goes to exiting
	// validators
	maxBalanceChurn := totalBalance / (p.MaxBalanceChurnQuotient * 2)

	balanceChurn := uint64(0)
	for idx, validator := range s.ValidatorRegistry {
		index := uint32(idx)

		// start validators if needed
		if validator.Status == StatusStarting && validator.Balance == p.DepositAmount*p.UnitsPerCoin && validator.FirstActiveEpoch <= s.EpochIndex {
			balanceChurn += s.GetEffectiveBalance(index, p)

			if balanceChurn > maxBalanceChurn {
				break
			}

			err := s.UpdateValidatorStatus(index, StatusActive, p)
			if err != nil {
				return err
			}
		}
	}

	balanceChurn = 0
	for idx, validator := range s.ValidatorRegistry {
		index := uint32(idx)

		if validator.Status == StatusActivePendingExit && validator.LastActiveEpoch <= s.EpochIndex {
			balanceChurn += s.GetEffectiveBalance(index, p)

			if balanceChurn > maxBalanceChurn {
				break
			}

			err := s.UpdateValidatorStatus(index, StatusExitedWithoutPenalty, p)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generateRandNumber(from chainhash.Hash, max uint32) uint64 {
	randaoBig := new(big.Int)
	randaoBig.SetBytes(from[:])

	numValidator := big.NewInt(int64(max + 1))

	return randaoBig.Mod(randaoBig, numValidator).Uint64()
}

// Shuffle shuffles validator using a RANDAO.
func Shuffle(randao chainhash.Hash, vals []uint32) []uint32 {
	nextProposers := make([]uint32, len(vals))
	copy(nextProposers, vals)

	for i := uint64(0); i < uint64(len(nextProposers)-1); i++ {
		j := i + generateRandNumber(randao, uint32(len(nextProposers))-uint32(i)-1)
		nextProposers[i], nextProposers[j] = nextProposers[j], nextProposers[i]
		randao = chainhash.HashH(randao[:])
	}

	return nextProposers
}

// DetermineNextProposers gets the next shuffling.
func DetermineNextProposers(randao chainhash.Hash, activeValidators []uint32, p *params.ChainParams) []uint32 {
	validatorsChosen := make(map[uint64]struct{})
	nextProposers := make([]uint32, p.EpochLength)

	for i := range nextProposers {
		found := true
		var val uint64

		for found {
			val = generateRandNumber(randao, uint32(len(activeValidators)-1))
			randao = chainhash.HashH(randao[:])
			_, found = validatorsChosen[val]
		}

		validatorsChosen[val] = struct{}{}
		nextProposers[i] = uint32(val)
		randao = chainhash.HashH(randao[:])
	}

	return nextProposers
}

// GetRecentBlockHash gets block hashes from the LatestBlockHashes array.
func (s *State) GetRecentBlockHash(slotToGet uint64, p *params.ChainParams) chainhash.Hash {
	if s.Slot-slotToGet >= p.LatestBlockRootsLength {
		return chainhash.Hash{}
	}

	return s.LatestBlockHashes[slotToGet%p.LatestBlockRootsLength]
}

type ReceiptType uint8

const (
	RewardMatchedFromEpoch ReceiptType = iota
	PenaltyMissingFromEpoch
	RewardMatchedToEpoch
	PenaltyMissingToEpoch
	RewardMatchedBeaconBlock
	PenaltyMissingBeaconBlock
	RewardIncludedVote
	RewardInclusionDistance
	PenaltyInactivityLeak
	PenaltyInactivityLeakNoVote
)

func (r ReceiptType) String() string {
	switch r {
	case RewardMatchedFromEpoch:
		return "voted for correct from epoch"
	case RewardMatchedToEpoch:
		return "voted for correct to epoch"
	case RewardMatchedBeaconBlock:
		return "voted for correct beacon"
	case RewardIncludedVote:
		return "included vote in proposal"
	case RewardInclusionDistance:
		return "inclusion distance reward"
	case PenaltyInactivityLeak:
		return "inactivity leak"
	case PenaltyInactivityLeakNoVote:
		return "did not vote with inactivity leak"
	case PenaltyMissingBeaconBlock:
		return "voted for wrong beacon block"
	case PenaltyMissingFromEpoch:
		return "voted for wrong from epoch"
	case PenaltyMissingToEpoch:
		return "voted for wrong to epoch"
	default:
		return fmt.Sprintf("invalid receipt type: %d", r)
	}
}

// EpochReceipt is a balance change carried our by an epoch transition.
type EpochReceipt struct {
	Type      ReceiptType
	Amount    int64
	Validator uint32
}

// Marshal serializes the struct to bytes
func (e *EpochReceipt) Marshal() ([]byte, error) {
	return e.Marshal()
}

// Unmarshal deserializes the struct from bytes
func (e *EpochReceipt) Unmarshal(b []byte) error {
	return e.Unmarshal(b)
}

func (e *EpochReceipt) String() string {
	if e.Amount > 0 {
		return fmt.Sprintf("Reward: Validator %d: %s for %f POLIS", e.Validator, e.Type, float64(e.Amount)/1000)
	} else {
		return fmt.Sprintf("Penalty: Validator %d: %s for %f POLIS", e.Validator, e.Type, float64(e.Amount)/1000)
	}
}

// GetTotalBalances gets the total balances of the state.
func (s *State) GetTotalBalances() uint64 {
	total := uint64(0)
	for _, v := range s.ValidatorRegistry {
		total += v.Balance
	}

	for _, v := range s.CoinsState.Balances {
		total += v
	}

	return total
}

// NextVoteEpoch increments the voting epoch, resets votes,
// and updates the state.
func (s *State) NextVoteEpoch(newState GovernanceState) {
	s.VoteEpoch++
	s.VoteEpochStartSlot = s.Slot
	s.CommunityVotes = make(map[chainhash.Hash]CommunityVoteData)
	s.ReplaceVotes = make(map[[20]byte]chainhash.Hash)
	s.VotingState = newState
}

// CheckForVoteTransitions tallies up votes and checks for any governance
// state transitions.
func (s *State) CheckForVoteTransitions(p *params.ChainParams) {
	switch s.VotingState {
	case GovernanceStateActive:
		// if it's active, we should check if we've accumulated enough votes
		// to start a community vote
		totalBalance := s.GetTotalBalances()
		votingBalance := uint64(0)
		for i := range s.ReplaceVotes {
			bal, ok := s.CoinsState.Balances[i]
			if !ok {
				continue
			}
			votingBalance += bal
		}

		if votingBalance*p.CommunityOverrideQuotient >= totalBalance {
			s.NextVoteEpoch(GovernanceStateVoting)
			for i := range s.CurrentManagers {
				s.ManagerReplacement.Set(uint(i))
			}
		}
	case GovernanceStateVoting:
		if s.VoteEpochStartSlot+p.VotingPeriodSlots <= s.Slot {
			// tally votes and choose next managers
			managerVotes := make(map[chainhash.Hash]uint64)

			for i, v := range s.ReplaceVotes {
				bal, ok := s.CoinsState.Balances[i]
				if !ok {
					continue
				}
				if _, ok := managerVotes[v]; ok {
					managerVotes[v] += bal
				} else {
					managerVotes[v] = bal
				}
			}

			bestBalance := uint64(0)
			bestManagers := s.CurrentManagers
			for i, v := range managerVotes {
				if v > bestBalance {
					voteData := s.CommunityVotes[i]

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
	epochsPerMonth := 30 * 24 * 60 * 60 / p.SlotDuration / p.EpochLength
	if s.LastPaidSlot/p.EpochLength+epochsPerMonth <= s.Slot {
		// 10% to 5/5 multisig
		// 10% to each

		totalBlockReward := p.BaseRewardPerBlock * 60 * 60 * 24 * 30 / p.SlotDuration
		perGroup := totalBlockReward / 10

		multipub := bls.PublicKeyHashesToMultisigHash(s.CurrentManagers, 5)
		s.CoinsState.Balances[multipub] += perGroup

		if len(s.CurrentManagers) != len(p.GovernancePercentages) {
			return
		}

		for group, address := range s.CurrentManagers {
			percent := p.GovernancePercentages[group]
			s.CoinsState.Balances[address] += perGroup * uint64(percent) / 100
		}

		s.LastPaidSlot = s.Slot
	}
}

// ProcessEpochTransition runs an epoch transition on the state.
func (s *State) ProcessEpochTransition(p *params.ChainParams, log *logger.Logger) ([]*EpochReceipt, error) {
	s.CheckForVoteTransitions(p)

	totalBalance := s.getActiveBalance(p)

	// These are voters who voted for a target of the previous epoch.
	previousEpochVoters := newVoterGroup()
	previousEpochVotersMatchingTargetHash := newVoterGroup()
	previousEpochVotersMatchingBeaconBlock := newVoterGroup()

	currentEpochVotersMatchingTarget := newVoterGroup()

	previousEpochBoundaryHash := chainhash.Hash{}
	if s.Slot >= 2*p.EpochLength {
		previousEpochBoundaryHash = s.GetRecentBlockHash(s.Slot-2*p.EpochLength-1, p)
	}

	epochBoundaryHash := chainhash.Hash{}
	if s.Slot >= p.EpochLength {
		epochBoundaryHash = s.GetRecentBlockHash(s.Slot-p.EpochLength-1, p)
	}

	// previousEpochVotersMap maps validator to their assigned vote
	previousEpochVotersMap := make(map[uint32]*AcceptedVoteInfo)

	for _, v := range s.PreviousEpochVotes {
		validatorIndices, err := s.GetVoteCommittee(v.Data.Slot, p)
		if err != nil {
			return nil, err
		}
		previousEpochVoters.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield, validatorIndices)
		actualBlockHash := s.GetRecentBlockHash(v.Data.Slot-1, p)
		if v.Data.BeaconBlockHash.IsEqual(&actualBlockHash) {
			previousEpochVotersMatchingBeaconBlock.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield, validatorIndices)
		}
		if v.Data.ToHash.IsEqual(&previousEpochBoundaryHash) {
			previousEpochVotersMatchingTargetHash.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield, validatorIndices)
		}
		for _, validatorIdx := range validatorIndices {
			previousEpochVotersMap[validatorIdx] = &v
		}
	}

	for _, v := range s.CurrentEpochVotes {
		validators, err := s.GetVoteCommittee(v.Data.Slot, p)
		if err != nil {
			return nil, err
		}

		if v.Data.ToHash.IsEqual(&epochBoundaryHash) {
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
		s.JustifiedEpochHash = s.GetRecentBlockHash(s.JustifiedEpoch*p.EpochLength, p)
	}

	if 3*currentEpochVotersMatchingTarget.totalBalance >= 2*totalBalance {
		s.JustificationBitfield |= 1 << 0
		s.JustifiedEpoch = s.EpochIndex
		s.JustifiedEpochHash = s.GetRecentBlockHash(s.JustifiedEpoch*p.EpochLength, p)
	}

	if (s.JustificationBitfield>>1)%4 == 3 && s.PreviousJustifiedEpoch == s.EpochIndex-2 { // 110 <- old previous justified would be
		s.FinalizedEpoch = s.PreviousJustifiedEpoch
	}

	if ((s.JustificationBitfield>>0)%8 == 7 && s.JustifiedEpoch == s.EpochIndex-1) || // 111
		((s.JustificationBitfield>>0)%4 == 3 && s.JustifiedEpoch == s.EpochIndex) {
		s.FinalizedEpoch = s.PreviousJustifiedEpoch
		s.JustifiedEpoch = s.EpochIndex
	}

	const numRewards = 5

	baseReward := func(index uint32) uint64 {
		return s.GetEffectiveBalance(index, p) * p.UnitsPerCoin * p.BaseRewardPerBlock * p.EpochLength / totalBalance / numRewards
	}

	receipts := make([]*EpochReceipt, 0)

	rewardValidator := func(index uint32, reward uint64, why ReceiptType) {
		s.ValidatorRegistry[index].Balance += reward
		receipts = append(receipts, &EpochReceipt{
			Validator: index,
			Amount:    int64(reward),
			Type:      why,
		})
	}

	penalizeValidator := func(index uint32, penalty uint64, why ReceiptType) {
		if s.ValidatorRegistry[index].FirstActiveEpoch+5 >= s.EpochIndex {
			return
		}
		s.ValidatorRegistry[index].Balance -= penalty
		receipts = append(receipts, &EpochReceipt{
			Validator: index,
			Amount:    -int64(penalty),
			Type:      why,
		})
	}

	if s.Slot >= 2*p.EpochLength {
		for index, validator := range s.ValidatorRegistry {
			idx := uint32(index)
			if !validator.IsActive() {
				continue
			}

			// votes matching source rewarded
			if previousEpochVoters.contains(idx) {
				reward := baseReward(idx)
				rewardValidator(idx, reward, RewardMatchedFromEpoch)
			} else {
				penalty := baseReward(idx)
				penalizeValidator(idx, penalty, PenaltyMissingFromEpoch)
			}

			// votes matching target rewarded
			if previousEpochVotersMatchingTargetHash.contains(idx) {
				reward := baseReward(idx)
				rewardValidator(idx, reward, RewardMatchedToEpoch)
			} else {
				penalty := baseReward(idx)
				penalizeValidator(idx, penalty, PenaltyMissingToEpoch)
			}

			// votes matching beacon block rewarded
			if previousEpochVotersMatchingBeaconBlock.contains(idx) {
				reward := baseReward(idx)
				rewardValidator(idx, reward, RewardMatchedBeaconBlock)
			} else {
				penalty := baseReward(idx)
				penalizeValidator(idx, penalty, PenaltyMissingBeaconBlock)
			}
		}

		// inclusion rewards
		proposerRewardInclusion := make(map[uint32]uint64)
		proposerRewardDistance := make(map[uint32]uint64)
		for voter := range previousEpochVoters.voters {
			vote := previousEpochVotersMap[voter]

			proposerIndex := vote.Proposer

			reward := baseReward(proposerIndex)
			if _, ok := proposerRewardInclusion[proposerIndex]; !ok {
				proposerRewardInclusion[proposerIndex] = 0
			}
			proposerRewardInclusion[proposerIndex] += reward

			inclusionDistance := vote.InclusionDelay

			reward = baseReward(proposerIndex) * p.MinAttestationInclusionDelay / inclusionDistance
			if _, ok := proposerRewardDistance[proposerIndex]; !ok {
				proposerRewardDistance[proposerIndex] = 0
			}
			proposerRewardDistance[proposerIndex] += reward
		}

		for validator, amount := range proposerRewardInclusion {
			rewardValidator(validator, amount, RewardIncludedVote)
		}
		for validator, amount := range proposerRewardDistance {
			rewardValidator(validator, amount, RewardInclusionDistance)
		}

		// Penalize all validators after 4 epochs and punish validators who did not
		// vote more severely.
		finalityDelay := s.EpochIndex - s.FinalizedEpoch
		if finalityDelay > 4 {
			for index := range s.ValidatorRegistry {
				idx := uint32(index)
				w := &s.ValidatorRegistry[idx]
				if !w.IsActive() {
					continue
				}

				penalty := baseReward(idx) * numRewards

				penalizeValidator(idx, penalty, PenaltyInactivityLeak)

				if !previousEpochVotersMatchingTargetHash.contains(idx) {
					penalty := s.GetEffectiveBalance(idx, p) * finalityDelay / p.InactivityPenaltyQuotient
					penalizeValidator(idx, penalty, PenaltyInactivityLeakNoVote)
				}
			}
		}
	}

	for index, validator := range s.ValidatorRegistry {
		if validator.IsActive() && validator.Balance < p.EjectionBalance*p.UnitsPerCoin {
			err := s.UpdateValidatorStatus(uint32(index), StatusExitedWithoutPenalty, p)
			if err != nil {
				return nil, err
			}
		}
	}

	s.EpochIndex = s.Slot / p.EpochLength

	if s.FinalizedEpoch > s.LatestValidatorRegistryChange {
		err := s.updateValidatorRegistry(p)
		if err != nil {
			return nil, err
		}

		s.LatestValidatorRegistryChange = s.EpochIndex
	}

	s.ProposerQueue = s.NextProposerQueue
	activeValidators := s.GetValidatorIndicesActiveAt(s.EpochIndex + 1)
	s.NextProposerQueue = DetermineNextProposers(s.RANDAO, activeValidators, p)

	s.PreviousEpochVoteAssignments = s.CurrentEpochVoteAssignments
	s.CurrentEpochVoteAssignments = Shuffle(s.RANDAO, activeValidators)

	copy(s.RANDAO[:], s.NextRANDAO[:])

	s.PreviousEpochVotes = s.CurrentEpochVotes
	s.CurrentEpochVotes = make([]AcceptedVoteInfo, 0)

	return receipts, nil
}
