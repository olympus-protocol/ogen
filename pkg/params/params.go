package params

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"math"
)

const (
	mayor = 0
	minor = 1
	patch = 2
)

var (
	VersionNumber = (mayor * 100000) + (minor * 1000) + (patch * 10)
	Version       = fmt.Sprintf("%d.%d.%d", mayor, minor, patch)
)

func ProtocolID(net string) protocol.ID {
	return protocol.ID("/ogen/" + net)
}

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
	MaxCoinProofsPerBlock        uint64
	MaxPartialExitsPerBlock      uint64
	MaxExecutionsPerBlock        uint64
	WhistleblowerRewardQuotient  uint64
	GovernancePercentages        []uint8
	MinVotingBalance             uint64
	CommunityOverrideQuotient    uint64
	VotingPeriodSlots            uint64
	InitialManagers              [][20]byte
	RendevouzStrings             map[int]string
	Relayers                     map[string]string
}

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
	MaxCoinProofsPerBlock:        128,
	MaxPartialExitsPerBlock:      128,
	MaxExecutionsPerBlock:        256,
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

var _, daoTest1Bytes, _ = bech32.Decode("tlpub1tppnrl6hv7gs2je6vrpa0xzrxyjuh32pnw4uua")
var _, daoTest2Bytes, _ = bech32.Decode("tlpub1nnpct659h9vwf2u3ty787txzy2kamwxp8zh42w")
var _, daoTest3Bytes, _ = bech32.Decode("tlpub1sfeshmckryr8q2z906qpym25npt9hmtn7049g5")
var _, daoTest4Bytes, _ = bech32.Decode("tlpub1xwl3hy063ml66x0qldekvjrnch8uakfzt8e3c7")
var _, daoTest5Bytes, _ = bech32.Decode("tlpub14whaq4vf7al3qswsrmxlyy6cfv62w4sfjdwtnr")

var daoTest1, daoTest2, daoTest3, daoTest4, daoTest5 [20]byte

func init() {
	copy(daoTest1[:], daoTest1Bytes)
	copy(daoTest2[:], daoTest2Bytes)
	copy(daoTest3[:], daoTest3Bytes)
	copy(daoTest4[:], daoTest4Bytes)
	copy(daoTest5[:], daoTest5Bytes)
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
	MaxCoinProofsPerBlock:        128,
	MaxPartialExitsPerBlock:      128,
	MaxExecutionsPerBlock:        256,
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
		daoTest1, // tlpub1tppnrl6hv7gs2je6vrpa0xzrxyjuh32pnw4uua | tlprv1pfmtg75uva0vdepd2w83ldlsrt3ayj7te8a95mrmuu8t0adylqvqxq9cz0
		daoTest2, // tlpub1nnpct659h9vwf2u3ty787txzy2kamwxp8zh42w | tlprv1zplv47hw33y2ks3mt2pqywxmhg3g7ct5u06074qxfvk73cvtz3qqnyvjnv
		daoTest3, // tlpub1sfeshmckryr8q2z906qpym25npt9hmtn7049g5 | tlprv1x4uewptrswvr7wxkhz2dzf7xvqejgghngqrpxemvnzhk0jr2cwtqd65adv
		daoTest4, // tlpub1xwl3hy063ml66x0qldekvjrnch8uakfzt8e3c7 | tlprv1ptllcm85dw3cgsdlvh6xrytyzhadtlqfg5e0vv39a6vdr3zgjk5qqm6xnp
		daoTest5, // tlpub14whaq4vf7al3qswsrmxlyy6cfv62w4sfjdwtnr | tlprv18nwwpaefghq8z568q6zs3tpqhlt7w8rygu6uksh9rtpcs9gue3jsx500gc
	},
	RendevouzStrings: map[int]string{
		0: "do_not_go_gentle_into_that_good_night",
	},
	Relayers: map[string]string{
		"cronos-1": "/ip4/134.122.28.156/tcp/25000/p2p/12D3KooWDv5BH9bQhv198TXGkXygNoXrEdvEfLkKL6C5eD3EAHvi",
		"cronos-2": "/ip4/159.65.233.200/tcp/25000/p2p/12D3KooWLDdEF8zAK7tQqDN23CmC4TFZqKeo2n95BJUBaJH69h5P",
	},
}

// DevNet are chain parameters used for the devnet.
var DevNet = ChainParams{
	Name:           "devnet",
	DefaultP2PPort: "26126",
	NetMagic:       111999,
	AccountPrefixes: AccountPrefixes{
		Public:   "dlpub",
		Private:  "dlprv",
		Multisig: "dlmul",
		Contract: "dlctr",
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
	MaxCoinProofsPerBlock:        128,
	MaxPartialExitsPerBlock:      128,
	MaxExecutionsPerBlock:        256,
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
		daoTest1, // tlpub1tppnrl6hv7gs2je6vrpa0xzrxyjuh32pnw4uua | tlprv1pfmtg75uva0vdepd2w83ldlsrt3ayj7te8a95mrmuu8t0adylqvqxq9cz0
		daoTest2, // tlpub1nnpct659h9vwf2u3ty787txzy2kamwxp8zh42w | tlprv1zplv47hw33y2ks3mt2pqywxmhg3g7ct5u06074qxfvk73cvtz3qqnyvjnv
		daoTest3, // tlpub1sfeshmckryr8q2z906qpym25npt9hmtn7049g5 | tlprv1x4uewptrswvr7wxkhz2dzf7xvqejgghngqrpxemvnzhk0jr2cwtqd65adv
		daoTest4, // tlpub1xwl3hy063ml66x0qldekvjrnch8uakfzt8e3c7 | tlprv1ptllcm85dw3cgsdlvh6xrytyzhadtlqfg5e0vv39a6vdr3zgjk5qqm6xnp
		daoTest5, // tlpub14whaq4vf7al3qswsrmxlyy6cfv62w4sfjdwtnr | tlprv18nwwpaefghq8z568q6zs3tpqhlt7w8rygu6uksh9rtpcs9gue3jsx500gc
	},
	RendevouzStrings: map[int]string{
		0: "do_not_go_gentle_into_that_good_night",
	},
	Relayers: map[string]string{
		"cronos-devnet-1": "/ip4/174.138.34.252/tcp/25000/p2p/12D3KooWDMH4ddviknJJocAJWAhGcTXXkhorkzvpdLAe2CdK8Fxd",
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
