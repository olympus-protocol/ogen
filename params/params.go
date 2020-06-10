package params

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// AddrPrefixes are prefixes used for addresses.
type AddrPrefixes struct {
	Public   string
	Private  string
	Multisig string
}

// ChainParams are parameters that are unique for the chain.
type ChainParams struct {
	Name                         string
	DefaultP2PPort               string
	GenesisHash                  chainhash.Hash
	AddrPrefix                   AddrPrefixes
	GovernanceBudgetQuotient     uint64
	EpochLength                  uint64
	EjectionBalance              uint64
	MaxBalanceChurnQuotient      uint64
	MaxVotesPerBlock             uint64
	MaxTxsPerBlock               uint64
	LatestBlockRootsLength       uint64
	MinAttestationInclusionDelay uint64
	DepositAmount                uint64
	BaseRewardPerBlock           uint64
	UnitsPerCoin                 uint64
	InactivityPenaltyQuotient    uint64
	IncluderRewardQuotient       uint64
	SlotDuration                 uint64
	MaxDepositsPerBlock          uint64
	MaxExitsPerBlock             uint64
	MaxRANDAOSlashingsPerBlock   uint64
	MaxProposerSlashingsPerBlock uint64
	MaxVoteSlashingsPerBlock     uint64
	WhistleblowerRewardQuotient  uint64
	GovernancePercentages        []uint8
	MinVotingBalance             uint64
	CommunityOverrideQuotient    uint64
	VotingPeriodSlots            uint64
	InitialManagers              [][20]byte

	ChainFileHash chainhash.Hash
	ChainFileURL  string
}

// Mainnet are chain parameters used for the main network.
var Mainnet = ChainParams{
	Name:           "mainnet",
	DefaultP2PPort: "24126",
	AddrPrefix: AddrPrefixes{
		Public:   "olpub",
		Private:  "olprv",
		Multisig: "olmul",
	},
	GovernanceBudgetQuotient:     5, // 20%
	BaseRewardPerBlock:           2600,
	IncluderRewardQuotient:       8,
	EpochLength:                  5,
	EjectionBalance:              1000, // POLIS
	MaxBalanceChurnQuotient:      8,
	MaxVotesPerBlock:             32,
	LatestBlockRootsLength:       64,
	MinAttestationInclusionDelay: 1,
	DepositAmount:                10000,
	UnitsPerCoin:                 1000,
	InactivityPenaltyQuotient:    17179869184,
	SlotDuration:                 60,
	MaxTxsPerBlock:               1000,
	MaxDepositsPerBlock:          32,
	MaxExitsPerBlock:             32,
	MaxRANDAOSlashingsPerBlock:   20,
	MaxProposerSlashingsPerBlock: 2,
	MaxVoteSlashingsPerBlock:     10,
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
}

var testnetChainFileHash, _ = chainhash.NewHashFromStr("b2d8f4ed146850d3b086c4a938179418bc30755ed9957a73f22e7c5a34e66ac2")

// TestNet are chain parameters used for the testnet.
var TestNet = ChainParams{
	Name:           "testnet",
	DefaultP2PPort: "24126",
	AddrPrefix: AddrPrefixes{
		Public:   "tlpub",
		Private:  "tlprv",
		Multisig: "tlmul",
	},
	GovernanceBudgetQuotient:     5, // 20%
	BaseRewardPerBlock:           2600,
	IncluderRewardQuotient:       8,
	EpochLength:                  5,
	EjectionBalance:              1000,
	MaxBalanceChurnQuotient:      32,
	MaxVotesPerBlock:             32,
	LatestBlockRootsLength:       64,
	MinAttestationInclusionDelay: 1,
	DepositAmount:                10000,
	UnitsPerCoin:                 1000,
	InactivityPenaltyQuotient:    17179869184,
	SlotDuration:                 2,
	MaxTxsPerBlock:               1000,
	ChainFileHash:                *testnetChainFileHash,
	ChainFileURL:                 "https://public.oly.tech/olympus/testnet/chain.json",
	MaxDepositsPerBlock:          32,
	MaxExitsPerBlock:             32,
	MaxRANDAOSlashingsPerBlock:   20,
	MaxProposerSlashingsPerBlock: 2,
	MaxVoteSlashingsPerBlock:     10,
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
