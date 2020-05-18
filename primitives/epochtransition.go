package primitives

import (
	"errors"
	"math/big"

	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/chainhash"
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

func (vg *voterGroup) addFromBitfield(registry []Worker, bitfield []uint8, validatorIndices []uint32) {
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
func (s *State) ExitValidator(index uint32, status WorkerStatus, p *params.ChainParams) error {
	validator := &s.ValidatorRegistry[index]
	prevStatus := validator.Status

	if prevStatus == StatusExitedWithPenalty {
		return nil
	}

	validator.Status = status

	if status == StatusExitedWithPenalty {
		// TODO: slashings - reward includer
		return nil
	}

	s.UtxoState.Balances[validator.PayeeAddress] += validator.Balance
	validator.Balance = 0

	return nil
}

// UpdateValidatorStatus moves a validator to a specific status.
func (s *State) UpdateValidatorStatus(index uint32, status WorkerStatus, p *params.ChainParams) error {
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
		if validator.Status == StatusStarting && validator.Balance == p.DepositAmount*p.UnitsPerCoin && validator.FirstActiveEpoch <= int64(s.EpochIndex) {
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

		if validator.Status == StatusActivePendingExit && validator.LastActiveEpoch <= int64(s.EpochIndex) {
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

// Shuffle shuffles workers using a RANDAO.
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

// ProcessEpochTransition runs an epoch transition on the state.
func (s *State) ProcessEpochTransition(p *params.ChainParams, log *logger.Logger) error {
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
		validatorIndices := s.GetVoteCommittee(v.Data.Slot, p)
		previousEpochVoters.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield, validatorIndices)
		actualBlockHash := s.GetRecentBlockHash(v.Data.Slot, p)
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
		validators := s.GetVoteCommittee(v.Data.Slot, p)

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

	baseRewardQuotient := p.BaseRewardQuotient * intSqrt(totalBalance/p.UnitsPerCoin)

	baseReward := func(index uint32) uint64 {
		return s.GetEffectiveBalance(index, p) / baseRewardQuotient / 5
	}

	if s.Slot >= 2*p.EpochLength {
		for index, validator := range s.ValidatorRegistry {
			idx := uint32(index)
			if !validator.IsActive() {
				continue
			}

			// votes matching source rewarded
			if previousEpochVoters.contains(idx) {
				reward := baseReward(idx) * previousEpochVoters.totalBalance / totalBalance
				s.ValidatorRegistry[idx].Balance += reward
			} else {
				penalty := baseReward(idx)
				s.ValidatorRegistry[idx].Balance -= penalty
			}

			// votes matching target rewarded
			if previousEpochVotersMatchingTargetHash.contains(idx) {
				reward := baseReward(idx) * previousEpochVoters.totalBalance / totalBalance
				s.ValidatorRegistry[idx].Balance += reward
			} else {
				penalty := baseReward(idx)
				s.ValidatorRegistry[idx].Balance -= penalty
			}

			// votes matching beacon block rewarded
			if previousEpochVotersMatchingBeaconBlock.contains(idx) {
				reward := baseReward(idx) * previousEpochVoters.totalBalance / totalBalance
				s.ValidatorRegistry[idx].Balance += reward
			} else {
				penalty := baseReward(idx)
				s.ValidatorRegistry[idx].Balance -= penalty
			}
		}

		// inclusion rewards
		for voter := range previousEpochVoters.voters {
			vote := previousEpochVotersMap[voter]

			proposerIndex := vote.Proposer

			reward := baseReward(proposerIndex) / p.IncluderRewardQuotient
			s.ValidatorRegistry[proposerIndex].Balance += reward

			inclusionDistance := vote.InclusionDelay

			reward = baseReward(voter) * p.MinAttestationInclusionDelay / inclusionDistance
			s.ValidatorRegistry[voter].Balance += reward
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

				w.Balance -= baseReward(idx) * 3
				if !previousEpochVotersMatchingTargetHash.contains(idx) {
					w.Balance -= s.GetEffectiveBalance(idx, p) * finalityDelay / p.InactivityPenaltyQuotient
				}
			}
		}
	}

	for index, validator := range s.ValidatorRegistry {
		if validator.IsActive() && validator.Balance < p.EjectionBalance*p.UnitsPerCoin {
			err := s.UpdateValidatorStatus(uint32(index), StatusExitedWithoutPenalty, p)
			if err != nil {
				return err
			}
		}
	}

	s.EpochIndex = s.Slot / p.EpochLength

	if s.FinalizedEpoch > s.LatestValidatorRegistryChange {
		err := s.updateValidatorRegistry(p)
		if err != nil {
			return err
		}

		s.LatestValidatorRegistryChange = s.EpochIndex
	}

	s.ProposerQueue = s.NextProposerQueue
	activeValidators := s.GetValidatorIndicesActiveAt(int64(s.EpochIndex + 1))
	s.NextProposerQueue = DetermineNextProposers(s.RANDAO, activeValidators, p)

	copy(s.PreviousEpochVoteAssignments, s.CurrentEpochVoteAssignments)

	copy(s.RANDAO[:], s.NextRANDAO[:])

	s.PreviousEpochVotes = s.CurrentEpochVotes
	s.CurrentEpochVotes = make([]AcceptedVoteInfo, 0)

	return nil
}
