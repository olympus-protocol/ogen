package chain

import (
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/amount"
)

func (ch *Blockchain) GetBlockReward(height uint32) amount.AmountType {
	baseReward := ch.params.BaseBlockReward
	cycles := height / ch.params.BlocksReductionCycle
	for i := uint32(0); i < cycles; i++ {
		baseReward -= baseReward * ch.params.BlockReductionPercentage
	}
	baseReward = baseReward - (baseReward * ch.params.GovernanceBudgetPercentage)
	return amount.AmountType(baseReward * 1e8)
}

func (ch *Blockchain) GetRewardBasedOnCollateral(height uint32, collateral p2p.OutPoint) amount.AmountType {
	return amount.AmountType(0)
}
