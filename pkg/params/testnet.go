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
	BaseRewardPerBlock:           18 * 1e7, // 1.8 POLIS
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
		"cronos-1-ipv4": "/ip4/128.199.244.76/tcp/25000/p2p/12D3KooWDvTjRxiQ4ysMd4GUv4EKbhtXix33QU4ANNFsMtah7AH1",
		"cronos-2-ipv4": "/ip4/128.199.244.102/tcp/25000/p2p/12D3KooWDAVSoS442h7fSFkoGRD6BJUXnijtcnoq7oyRoT9cVu9v",
	},
}
