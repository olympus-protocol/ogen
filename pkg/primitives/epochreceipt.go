package primitives

import "fmt"

const (
	RewardMatchedFromEpoch uint64 = iota
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

// EpochReceipt is a balance change carried our by an epoch transition.
type EpochReceipt struct {
	Type      uint64
	Amount    uint64
	Validator uint64
}

func (e EpochReceipt) TypeString() string {
	switch e.Type {
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
		return fmt.Sprintf("invalid receipt type: %d", e.Type)
	}
}

func (e *EpochReceipt) String() string {
	if e.Amount > 0 {
		return fmt.Sprintf("Reward: Validator %d: %s for %f POLIS", e.Validator, e.TypeString(), float64(e.Amount)/1000)
	} else {
		return fmt.Sprintf("Penalty: Validator %d: %s for %f POLIS", e.Validator, e.TypeString(), float64(e.Amount)/1000)
	}
}
