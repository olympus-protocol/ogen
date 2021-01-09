package params

// MainNet are chain parameters used for the main network.
var MainNet = ChainParams{
	Name:           "mainnet",
	DefaultP2PPort: "24126",
	NetMagic:       333999,
	AccountPrefixes: AccountPrefixes{
		Public:   "olpub",
		Private:  "olprv",
		Multisig: "olmul",
		Contract: "olctr",
	},
	GovernanceBudgetQuotient:     5,        // 20%
	BaseRewardPerBlock:           18 * 1e7, // 1.8 POLIS
	IncluderRewardQuotient:       8,
	EpochLength:                  5,
	EjectionBalance:              95, // POLIS
	MaxBalanceChurnQuotient:      8,
	LatestBlockRootsLength:       64,
	MinAttestationInclusionDelay: 1,
	DepositAmount:                100,
	UnitsPerCoin:                 100000000,
	InactivityPenaltyQuotient:    17179869184,
	SlotDuration:                 30,
	WhistleblowerRewardQuotient:  2, // Validator loses half their deposit
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
	RendevouzStrings: map[int]string{
		0: "do_not_go_gentle_into_that_good_night",
	},
}
