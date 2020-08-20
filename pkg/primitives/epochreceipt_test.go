package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEpochReceipt(t *testing.T) {
	e := primitives.EpochReceipt{
		Type:      primitives.RewardMatchedFromEpoch,
		Amount:    100,
		Validator: 50,
	}

	str := e.String()
	assert.Equal(t,"Reward: Validator 50: voted for correct from epoch for 0.100000 POLIS", str)
	e.Amount = -100
	str = e.String()
	assert.Equal(t,  "Penalty: Validator 50: voted for correct from epoch for -0.100000 POLIS", str)

	assert.Equal(t, "voted for correct from epoch", e.TypeString())
	e.Type = primitives.PenaltyMissingFromEpoch
	assert.Equal(t, e.TypeString(), "voted for wrong from epoch", e.TypeString())
	e.Type = primitives.RewardMatchedToEpoch
	assert.Equal(t, e.TypeString(), "voted for correct to epoch", e.TypeString())
	e.Type = primitives.PenaltyMissingToEpoch
	assert.Equal(t, e.TypeString(), "voted for wrong to epoch", e.TypeString())
	e.Type = primitives.RewardMatchedBeaconBlock
	assert.Equal(t, e.TypeString(), "voted for correct beacon", e.TypeString())
	e.Type = primitives.PenaltyMissingFromEpoch
	assert.Equal(t, e.TypeString(), "voted for wrong from epoch", e.TypeString())
	e.Type = primitives.PenaltyMissingBeaconBlock
	assert.Equal(t, e.TypeString(), "voted for wrong beacon block", e.TypeString())
	e.Type = primitives.RewardIncludedVote
	assert.Equal(t, e.TypeString(), "included vote in proposal", e.TypeString())
	e.Type = primitives.RewardInclusionDistance
	assert.Equal(t, e.TypeString(), "inclusion distance reward", e.TypeString())
	e.Type = primitives.PenaltyInactivityLeak
	assert.Equal(t, e.TypeString(), "inactivity leak", e.TypeString())
	e.Type = primitives.PenaltyInactivityLeakNoVote
	assert.Equal(t, e.TypeString(), "did not vote with inactivity leak", e.TypeString())
	e.Type = 10
	assert.Equal(t, e.TypeString(), "invalid receipt type: 10", e.TypeString())
}
