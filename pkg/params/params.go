package params

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"math"
)

const (
	mayor = 0
	minor = 1
	patch = 0
)

var (
	VersionNumber = (mayor * 100000) + (minor * 1000) + (patch * 10)
	Version       = fmt.Sprintf("%d.%d.%d", mayor, minor, patch)

	SyncProtocolID = protocol.ID("/ogen/sync/" + Version)
)

// AccountPrefixes are prefixes used for account bech32 encoding.
type AccountPrefixes struct {
	Public   string
	Private  string
	Multisig string
	Contract string
}

// ChainParams are parameters that are unique for the chain.
type ChainParams struct {
	Name                         string
	DefaultP2PPort               string
	GenesisHash                  chainhash.Hash
	AccountPrefixes              AccountPrefixes
	NetMagic                     uint32
	GovernanceBudgetQuotient     uint64
	EpochLength                  uint64
	EjectionBalance              uint64
	MaxBalanceChurnQuotient      uint64
	MaxVotesPerBlock             uint64
	MaxTxsPerBlock               uint64
	MaxTxsMultiPerBlock          uint64
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
	MaxGovernanceVotesPerBlock   uint64
	WhistleblowerRewardQuotient  uint64
	GovernancePercentages        []uint8
	MinVotingBalance             uint64
	CommunityOverrideQuotient    uint64
	VotingPeriodSlots            uint64
	InitialManagers              [][20]byte
	RendevouzStrings             map[int]string
	Relayers                     map[string]string
}

// Mainnet are chain parameters used for the main network.
var Mainnet = ChainParams{
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
	BaseRewardPerBlock:           26 * 1e7, // 2.6 POLIS
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
	MaxVotesPerBlock:             32,
	MaxTxsPerBlock:               5000,
	MaxTxsMultiPerBlock:          128,
	MaxDepositsPerBlock:          128,
	MaxExitsPerBlock:             128,
	MaxRANDAOSlashingsPerBlock:   20,
	MaxProposerSlashingsPerBlock: 2,
	MaxVoteSlashingsPerBlock:     10,
	MaxGovernanceVotesPerBlock:   128,
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
	MaxVotesPerBlock:             32,
	MaxTxsPerBlock:               5000,
	MaxTxsMultiPerBlock:          128,
	MaxDepositsPerBlock:          128,
	MaxExitsPerBlock:             128,
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
		{252, 94, 117, 132, 63, 93, 202, 26, 36, 23, 195, 26, 169, 95, 74, 147, 72, 184, 66, 20},        // tlpub1l308tpplth9p5fqhcvd2jh62jdytsss54nt6d4
		{192, 13, 158, 167, 115, 190, 56, 51, 43, 11, 156, 43, 27, 145, 143, 61, 40, 209, 114, 238},     // tlpub1cqxeafmnhcurx2ctns43hyv0855dzuhwnllx6w
		{88, 192, 115, 125, 142, 126, 244, 13, 253, 225, 139, 36, 184, 34, 71, 31, 69, 205, 216, 125},   // tlpub1trq8xlvw0m6qml0p3vjtsgj8razumkrawvwzza
		{143, 17, 152, 250, 184, 122, 141, 208, 109, 72, 148, 187, 248, 89, 83, 127, 113, 217, 23, 144}, // tlpub13uge374c02xaqm2gjjalsk2n0acaj9uswmr687
		{162, 207, 33, 52, 96, 81, 17, 131, 72, 175, 180, 222, 125, 41, 3, 108, 43, 47, 231, 7},         // tlpub15t8jzdrq2ygcxj90kn0862grds4jlec8tjcg6j
	},
	RendevouzStrings: map[int]string{
		0: "do_not_go_gentle_into_that_good_night",
	},
	Relayers: map[string]string{
		"cronos-1": "/ip4/134.122.28.156/tcp/25000/p2p/12D3KooWDbvbYMzEvRi5CECVHawadaaZLich93GRhaCRRxQHFVcN",
		"cronos-2": "/ip4/159.65.233.200/tcp/25000/p2p/12D3KooWBdwHS4aLQmfzhddfsT3aV9YA8Cy47VWx8ckoYkC8Vwb4",
	},
}

// GetRendevouzString is a function to return a rendevouz string for a certain version range
// to make sure peers find each other depending on their version.
func (p *ChainParams) GetRendevouzString() string {
	ver := VersionNumber
	var selectedIndex int
	var diffSelected int
	for n := range p.RendevouzStrings {
		diff := int(math.Abs(float64(ver - n)))
		if diff < diffSelected {
			selectedIndex = n
			diffSelected = diff
		}
	}
	return p.RendevouzStrings[selectedIndex]
}
