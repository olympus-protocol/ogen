package params

// TestNet are chain parameters used for the testnet.
var TestNet = ChainParams{
	Name:           "testnet",
	DefaultP2PPort: "25126",
	NetMagic:       222999,
	AccountPrefixes: AccountPrefixes{
		Public:   "tlpub",
		Private:  "tlprv",
		Multisig: "tlmul",
		Contract: "tlctr",
	},
	GovernanceBudgetQuotient:     5,        // 20%
	BaseRewardPerBlock:           26 * 1e7, // 2.6 POLIS
	ProofsMerkleRoot:             merkleRootHashTestNet,
	IncluderRewardQuotient:       8,
	EpochLength:                  5,
	EjectionBalance:              95,
	MaxBalanceChurnQuotient:      32,
	LatestBlockRootsLength:       64,
	MinAttestationInclusionDelay: 1,
	DepositAmount:                100,
	UnitsPerCoin:                 100000000,
	InactivityPenaltyQuotient:    17179869184,
	SlotDuration:                 30,
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
	RendevouzStrings: map[int]string{
		0: "do_not_go_gentle_into_that_good_night",
	},
	Relayers: map[string]string{
		"cronos-1-ipv4": "/ip4/198.199.88.226/tcp/25000/p2p/12D3KooWPxEqRMkvQN7eCdkEuaxN941u8PN5yKwfPbYVe1ujpLf6",
		"cronos-1-ipv6": "/ip6/2604:a880:400:d0::17ba:5001/tcp/25000/p2p/12D3KooWPxEqRMkvQN7eCdkEuaxN941u8PN5yKwfPbYVe1ujpLf6",
		"cronos-2-ipv4": "/ip4/159.203.176.202/tcp/25000/p2p/12D3KooWFGYWT99jkRpu2fuMFYir8xjvUszWMgJb2vv1iK5xKEm8",
		"cronos-2-ipv6": "/ip6/2604:a880:400:d0::1871:e001/tcp/25000/p2p/12D3KooWFGYWT99jkRpu2fuMFYir8xjvUszWMgJb2vv1iK5xKEm8",
	},
}
