package primitives_test

import (
	"github.com/magiconair/properties/assert"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"testing"
)

func TestEpochReceipt(t *testing.T) {
	e := primitives.EpochReceipt{
		Type:      primitives.RewardMatchedFromEpoch,
		Amount:    100,
		Validator: 50,
	}

	str := e.String()
	assert.Equal(t, str, "Reward: Validator 50: voted for correct from epoch for 0.100000 POLIS")
	e.Amount = -100
	str = e.String()
	assert.Equal(t, str, "Penalty: Validator 50: voted for correct from epoch for -0.100000 POLIS")

	assert.Equal(t, e.TypeString(), "voted for correct from epoch")
	e.Type = primitives.PenaltyMissingFromEpoch
	assert.Equal(t, e.TypeString(), "voted for wrong from epoch")
	e.Type = primitives.RewardMatchedToEpoch
	assert.Equal(t, e.TypeString(), "voted for correct to epoch")
	e.Type = primitives.PenaltyMissingToEpoch
	assert.Equal(t, e.TypeString(), "voted for wrong to epoch")
	e.Type = primitives.RewardMatchedBeaconBlock
	assert.Equal(t, e.TypeString(), "voted for correct beacon")
	e.Type = primitives.PenaltyMissingFromEpoch
	assert.Equal(t, e.TypeString(), "voted for wrong from epoch")
	e.Type = primitives.PenaltyMissingBeaconBlock
	assert.Equal(t, e.TypeString(), "voted for wrong beacon block")
	e.Type = primitives.RewardIncludedVote
	assert.Equal(t, e.TypeString(), "included vote in proposal")
	e.Type = primitives.RewardInclusionDistance
	assert.Equal(t, e.TypeString(), "inclusion distance reward")
	e.Type = primitives.PenaltyInactivityLeak
	assert.Equal(t, e.TypeString(), "inactivity leak")
	e.Type = primitives.PenaltyInactivityLeakNoVote
	assert.Equal(t, e.TypeString(), "did not vote with inactivity leak")
	e.Type = 10
	assert.Equal(t, e.TypeString(), "invalid receipt type: 10")
}
