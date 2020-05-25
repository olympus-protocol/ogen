package params

import (
	"github.com/olympus-protocol/ogen/utils/amount"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type AddrPrefixes struct {
	Public  string
	Private string
}

type ChainParams struct {
	Name                         string
	DefaultP2PPort               string
	GenesisHash                  chainhash.Hash
	AddrPrefix                   AddrPrefixes
	LastPreWorkersBlock          uint32
	PreWorkersPubKeyHash         string
	BlocksReductionCycle         uint32
	SuperBlockCycle              uint32
	SuperBlockStartHeight        uint32
	GovernanceBudgetPercentage   float64
	ProfitSharingCycle           uint32
	ProfitSharingStartCycle      uint32
	GovernanceProposalFee        amount.AmountType
	BlockReductionPercentage     float64
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

	ChainFileHash chainhash.Hash
	ChainFileURL  string
}

var NetworkNames = map[string]string{
	"mainnet": "Main Network",
	"test":    "Test Network",
}

var Mainnet = ChainParams{
	Name:           "polis",
	DefaultP2PPort: "24126",
	AddrPrefix: AddrPrefixes{
		Public:  "olpub",
		Private: "olprv",
	},
	LastPreWorkersBlock:          500,
	PreWorkersPubKeyHash:         "olpub12vjdayxm6eygqkxrtyvt0jnjxn8965wflynmf4d899pnkzp9glmslqcvce",
	BlocksReductionCycle:         262800, // 1 year
	SuperBlockCycle:              21600,  // 1 month
	SuperBlockStartHeight:        0,      // TODO define
	ProfitSharingCycle:           21600,  // 1 month
	ProfitSharingStartCycle:      0,      // TODO define
	GovernanceBudgetPercentage:   0.2,    // 20%
	BlockReductionPercentage:     0.2,    // 20%
	BaseRewardPerBlock:           2600,
	IncluderRewardQuotient:       8,
	GovernanceProposalFee:        amount.AmountType(50), // 50 POLIS
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
}

var testnetChainFileHash, _ = chainhash.NewHashFromStr("15f838a029028288ae8c5a5d07a2e6a4a5608d08fa3937f75c295d62f6fb30aa")

var TestNet = ChainParams{
	Name:                "test",
	DefaultP2PPort:      "24126",
	LastPreWorkersBlock: 10,
	AddrPrefix: AddrPrefixes{
		Public:  "tlpub",
		Private: "tlprv",
	},
	PreWorkersPubKeyHash:         "1HWfiw9Lbg2vh8A1sZDsp5BVLHeW41V13R", // 5JbK2h1P7BQTmwJgCPRonJzCqRMNpFPvsAPTwrHBdT7DmEzzsUK
	BlocksReductionCycle:         259200,                               // 6 months
	SuperBlockCycle:              1440,                                 // 1 day
	GovernanceBudgetPercentage:   0.2,                                  // 20%
	BlockReductionPercentage:     0.2,                                  // 20%
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
}
