package testdata

import (
	"github.com/olympus-protocol/ogen/pkg/params"
)

// TestParams network parameters for test chains.
var TestParams = params.ChainParams{
	Name:           "testnet",
	DefaultP2PPort: "25126",
	NetMagic:       111999,
	AccountPrefixes: params.AccountPrefixes{
		Public:   "itpub",
		Private:  "itprv",
		Multisig: "itmul",
		Contract: "itctr",
	},
	GovernanceBudgetQuotient:     5,        // 20%
	BaseRewardPerBlock:           26 * 1e7, // 2.6 POLIS
	IncluderRewardQuotient:       8,
	EpochLength:                  5,
	EjectionBalance:              95,
	MaxBalanceChurnQuotient:      32,
	LatestBlockRootsLength:       64,
	MinAttestationInclusionDelay: 1,
	DepositAmount:                100,
	UnitsPerCoin:                 100000000,
	InactivityPenaltyQuotient:    17179869184,
	SlotDuration:                 6,
	WhistleblowerRewardQuotient:  2,
	GovernancePercentages: []uint8{
		30, // tech
		10, // community
		20, // business
		20, // marketing
		20, // adoption
	},
	MinVotingBalance:          100,
	CommunityOverrideQuotient: 3,
	VotingPeriodSlots:         20160, // minutes in a week
	InitialManagers: [][20]byte{
		{},
		{},
		{},
		{},
		{},
	},
}
