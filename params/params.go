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

var testnetChainFileHash, _ = chainhash.NewHashFromStr("594d573a59827494fe78dcdfb432056ef81c69089f8fe776b5ffa2a0be7056d8")

// TestNet are chain parameters used for the testnet.
var TestNet = ChainParams{
	Name:           "testnet",
	DefaultP2PPort: "25126",
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
		{248, 16, 57, 125, 50, 192, 139, 189, 85, 90, 234, 131, 70, 18, 1, 43, 169, 114, 118, 177},  // tlpub1lqgrjlfjcz9m6426a2p5vysp9w5hya4372zwfl
		{13, 211, 240, 85, 95, 28, 4, 163, 48, 250, 130, 93, 18, 165, 63, 200, 2, 145, 174, 210},    // tlpub1phflq42lrsz2xv86sfw39ffleqpfrtkj6cykgs
		{95, 113, 185, 142, 139, 25, 233, 68, 13, 73, 241, 67, 55, 90, 155, 74, 57, 60, 143, 108},   // tlpub1tacmnr5tr855gr2f79pnwk5mfgunermv0ac08x
		{245, 18, 154, 38, 224, 21, 192, 106, 93, 106, 34, 1, 1, 110, 22, 227, 18, 117, 239, 38},    // tlpub175ff5fhqzhqx5ht2ygqszmskuvf8tmex84y5fl
		{184, 245, 90, 21, 89, 48, 7, 42, 187, 125, 223, 117, 51, 231, 119, 110, 174, 102, 68, 228}, // tlpub1hr64592exqrj4wmama6n8emhd6hxv38yac2shc
	},
}
