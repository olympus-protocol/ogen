package params

import (
	"encoding/hex"
	"fmt"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"math"
)

const (
	mayor = 0
	minor = 2
	patch = 0
)

var (
	VersionNumber = (mayor * 100000) + (minor * 1000) + (patch * 10)
	Version       = fmt.Sprintf("%d.%d.%d", mayor, minor, patch)
)

func ProtocolID(net string) protocol.ID {
	return protocol.ID("/ogen/" + net)
}

var merkleRootHashTestNet [32]byte

func init() {
	hashBytes, _ := hex.DecodeString("ef801c6398f121afafca8cf7b5a121e26d42d9b05f6711efe0a7687b670fcc7f") //  PolisBlockchain "height": 750711
	copy(merkleRootHashTestNet[:], hashBytes)
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

	/* Main Params */

	// Name is the common name of the network
	Name string
	// DefaultP2PPort is the default P2P port on which outbound/inbound connections are handled
	DefaultP2PPort string
	// GenesisHash is the hash of the genesis block
	GenesisHash chainhash.Hash
	// AccountPrefixes are the prefixes for bech32 accounts generator.
	AccountPrefixes AccountPrefixes
	// NetMagic is a number to serve as an  ID for Olympus specific messages
	NetMagic uint32
	// UnitsPerCoin is the amount of decimals used for coins.
	UnitsPerCoin uint64
	// RendevouzStrings are strings versioned for the DHT Peer relayer
	RendevouzStrings map[int]string
	// Relayers are the initial seeds to find peers.
	Relayers map[string]string
	// ProofsMerkleRoot is the merkle root to verify migration CoinProofs
	ProofsMerkleRoot chainhash.Hash

	/* Epochs & Slots related params */

	// EpochLength the amount of slots on an epoch.
	EpochLength uint64
	// SlotDuration is the amount of seconds for a slot.
	SlotDuration                 uint64
	MaxBalanceChurnQuotient      uint64
	LatestBlockRootsLength       uint64
	MinAttestationInclusionDelay uint64
	BaseRewardPerBlock           uint64

	/* Validators related params */

	// EjectionBalance the minimum validator balance to be exited from the network.
	EjectionBalance uint64
	// DepositAmount is the amount of coins that should be locked for a deposit.
	DepositAmount               uint64
	InactivityPenaltyQuotient   uint64
	IncluderRewardQuotient      uint64
	WhistleblowerRewardQuotient uint64

	/* Governance params */

	GovernanceBudgetQuotient  uint64
	GovernancePercentages     []uint8
	InitialManagers           [][20]byte
	VotingPeriodSlots         uint64
	MinVotingBalance          uint64
	CommunityOverrideQuotient uint64
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
