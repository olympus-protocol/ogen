package testdata

import (
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var MockGovernanceState = primitives.Governance{
	ReplaceVotes: map[[20]byte]chainhash.Hash{
		{0x1, 0x2, 0x3}: *Hash,
	},
	CommunityVotes: map[chainhash.Hash]primitives.CommunityVoteData{
		*Hash: CommunityVoteData,
	},
}

var MockCoinState = primitives.CoinsState{
	Balances: map[[20]byte]uint64{
		{0x0, 0x1, 0x2}: 400 * 1000000, // 400k coins
	},
	Nonces: make(map[[20]byte]uint64),
}

var MockState = primitives.State{
	CoinsState:                    MockCoinState,
	ValidatorRegistry:             []primitives.Validator{Validator, Validator, Validator, Validator},
	LatestValidatorRegistryChange: 1000,
	RANDAO:                        *Hash,
	NextRANDAO:                    *Hash,
	Slot:                          123,
	EpochIndex:                    123,
	ProposerQueue:                 []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9},
	PreviousEpochVoteAssignments:  []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9},
	CurrentEpochVoteAssignments:   []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9},
	NextProposerQueue:             []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9},
	JustificationBitfield:         123,
	FinalizedEpoch:                123,
	LatestBlockHashes:             []chainhash.Hash{*Hash, *Hash, *Hash, *Hash, *Hash},
	JustifiedEpoch:                15,
	JustifiedEpochHash:            *Hash,
	CurrentEpochVotes:             []primitives.AcceptedVoteInfo{AcceptedVoteInfo, AcceptedVoteInfo},
	PreviousJustifiedEpoch:        14,
	PreviousJustifiedEpochHash:    *Hash,
	PreviousEpochVotes:            []primitives.AcceptedVoteInfo{AcceptedVoteInfo, AcceptedVoteInfo},
	CurrentManagers:               [][20]byte{},
	ManagerReplacement:            bitfield.NewBitfield(15),
	VoteEpoch:                     15,
	VoteEpochStartSlot:            16,
	VotingState:                   1,
	LastPaidSlot:                  16,
}

func init() {

}
