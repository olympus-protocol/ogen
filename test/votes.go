package testdata

import (
	"github.com/olympus-protocol/ogen/primitives"
)

var VoteData = primitives.VoteData{
	Slot:      1,
	FromEpoch: 2,
	FromHash:  [32]byte{3},
	ToEpoch:   4,
	ToHash:    [32]byte{5},
}

var AcceptedVoteInfo = primitives.AcceptedVoteInfo{
	Data:                  VoteData,
	ParticipationBitfield: []uint8{6, 7},
	Proposer:              8,
	InclusionDelay:        9,
}

var SingleValidatorVote = primitives.SingleValidatorVote{
	Data:   VoteData,
	Sig:    sigB,
	Offset: 333,
	OutOf:  444,
}

var MultiValidatorVote = primitives.MultiValidatorVote{
	Data:                  VoteData,
	Sig:                   sigB,
	ParticipationBitfield: []byte{1, 2, 3, 4},
}
