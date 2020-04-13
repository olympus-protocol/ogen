package primitives

import (
	"errors"
	"math/big"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func (s *State) getEffectiveBalance(index uint32, p *params.ChainParams) uint64 {
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

func (vg *voterGroup) addFromBitfield(validators []Worker, bitfield []uint8) {
	for i, b := range bitfield {
		for j := 0; j < 8; j++ {
			val := i*8 + j
			if b&(1<<j) > 0 && val < len(validators) {
				vg.add(uint32(val), validators[val].Balance)
			}
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
	}

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
		if validator.Status == StatusStarting && validator.Balance == p.DepositAmount {
			balanceChurn += s.getEffectiveBalance(index, p)

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

		if validator.Status == StatusActivePendingExit {
			balanceChurn += s.getEffectiveBalance(index, p)

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

	numValidator := big.NewInt(int64(max))

	return randaoBig.Mod(randaoBig, numValidator).Uint64()
}

// DetermineNextProposers gets the next shuffling.
func DetermineNextProposers(randao chainhash.Hash, registry []Worker, p *params.ChainParams) []uint32 {
	validatorsChosen := make(map[uint64]struct{})
	nextProposers := make([]uint32, p.EpochLength)

	for i := range nextProposers {
		found := true
		var val uint64

		for found {
			val = generateRandNumber(randao, uint32(len(registry)))
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

func (s *State) ProcessEpochTransition(p *params.ChainParams) error {
	totalBalance := s.getActiveBalance(p)

	log.Infof("processing epoch transition with current votes: %d, prev votes: %d", len(s.CurrentEpochVotes), len(s.PreviousEpochVotes))

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
		epochBoundaryHash = s.GetRecentBlockHash(s.Slot-p.EpochLength, p)
	}

	previousEpochVotersMap := make(map[uint32]*AcceptedVoteInfo)

	for _, v := range s.PreviousEpochVotes {
		previousEpochVoters.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield)
		log.Infof("hash: %s, balance: %d, %v", v.Data.Hash(), previousEpochVoters.totalBalance, v.ParticipationBitfield)
		actualBlockHash := s.GetRecentBlockHash(v.Data.Slot, p)
		if v.Data.BeaconBlockHash.IsEqual(&actualBlockHash) {
			previousEpochVotersMatchingBeaconBlock.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield)
		}
		log.Infof("tohash: %s, actual: %s", v.Data.ToHash, previousEpochBoundaryHash)
		if v.Data.ToHash.IsEqual(&previousEpochBoundaryHash) {
			previousEpochVotersMatchingTargetHash.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield)
		}
		for i := range v.ParticipationBitfield {
			for j := 0; j < 8; j++ {
				val := uint32(i*8 + j)
				previousEpochVotersMap[val] = &v
			}
		}
	}

	for _, v := range s.CurrentEpochVotes {
		if v.Data.ToHash.IsEqual(&epochBoundaryHash) {
			currentEpochVotersMatchingTarget.addFromBitfield(s.ValidatorRegistry, v.ParticipationBitfield)
		}
	}

	oldPreviousJustifiedEpoch := s.PreviousJustifiedEpoch
	s.PreviousJustifiedEpoch = s.JustifiedEpoch
	s.PreviousJustifiedEpochHash = s.JustifiedEpochHash
	s.JustificationBitfield <<= 1

	// >2/3 voted with target of the previous epoch
	if 3*previousEpochVotersMatchingTargetHash.totalBalance >= 2*totalBalance {
		s.JustificationBitfield |= 1 << 1 // mark
	}

	if 3*currentEpochVotersMatchingTarget.totalBalance >= 2*totalBalance {
		s.JustificationBitfield |= 1 << 0
	}

	if ((s.JustificationBitfield>>1)%8 == 7 && s.PreviousJustifiedEpoch == s.EpochIndex-2) || // 1110
		((s.JustificationBitfield>>1)%4 == 3 && s.PreviousJustifiedEpoch == s.EpochIndex-1) { // 110 <- old previous justified would be
		s.FinalizedEpoch = oldPreviousJustifiedEpoch
	}

	if ((s.JustificationBitfield>>0)%8 == 7 && s.JustifiedEpoch == s.EpochIndex-1) || // 111
		((s.JustificationBitfield>>0)%4 == 3 && s.JustifiedEpoch == s.EpochIndex) {
		s.FinalizedEpoch = s.PreviousJustifiedEpoch
	}

	baseRewardQuotient := p.BaseRewardQuotient * intSqrt(totalBalance/p.UnitsPerCoin)

	baseReward := func(index uint32) uint64 {
		return s.getEffectiveBalance(index, p) / baseRewardQuotient / 5
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
					w.Balance -= s.getEffectiveBalance(idx, p) * finalityDelay / p.InactivityPenaltyQuotient
				}
			}
		}
	}

	for index, validator := range s.ValidatorRegistry {
		if validator.IsActive() && validator.Balance < p.EjectionBalance {
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
	}

	s.ProposerQueue = s.NextProposerQueue
	s.NextProposerQueue = DetermineNextProposers(s.RANDAO, s.ValidatorRegistry, p)

	copy(s.RANDAO[:], s.NextRANDAO[:])

	s.PreviousEpochVotes = s.CurrentEpochVotes
	s.CurrentEpochVotes = make([]AcceptedVoteInfo, 0)

	return nil
}
