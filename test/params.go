package testdata

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/params"
)

// KeystorePass is a simple string password for keystore tests
var KeystorePass = "test_pass"

// Node1Folder is the folder where node 1 stores its data
var Node1Folder = "./data_node1"

// Node2Folder is the folder where node 2 stores its data
var Node2Folder = "./data_node2"

// Node3Folder is the folder where node 3 stores its data
var Node3Folder = "./data_node3"

// PremineAddr a random bls key to use as premine addr
var PremineAddr = bls.RandKey()

// Conf are the test configuration flags
var Conf = config.Config{
	NetworkName:  "integration tests net",
	AddNodes:     []peer.AddrInfo{},
	MaxPeers:     10,
	Port:         "24126",
	RPCProxy:     false,
	RPCProxyPort: "8080",
	RPCPort:      "24130",
	RPCWallet:    false,
	Debug:        true,
	LogFile:      false,
	Pprof:        true,
}

// IntTestParams network parameters for test chains.
var IntTestParams = params.ChainParams{
	Name:           "testnet",
	DefaultP2PPort: "25126",
	AccountPrefixes: params.AccountPrefixes{
		Public:   "itpub",
		Private:  "itprv",
		Multisig: "itmul",
		Contract: "itctr",
	},
	GovernanceBudgetQuotient:     5, // 20%
	BaseRewardPerBlock:           2600,
	IncluderRewardQuotient:       8,
	EpochLength:                  5,
	EjectionBalance:              95,
	MaxBalanceChurnQuotient:      32,
	LatestBlockRootsLength:       64,
	MinAttestationInclusionDelay: 1,
	DepositAmount:                100,
	UnitsPerCoin:                 1000,
	InactivityPenaltyQuotient:    17179869184,
	SlotDuration:                 1,
	MaxVotesPerBlock:             2048,
	MaxTxsPerBlock:               5000,
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
}
