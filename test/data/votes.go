package testdata

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/primitives"
)

var sig = bls.NewAggregateSignature()

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
	Data:      VoteData,
	Signature: *sig,
	Offset:    333,
	OutOf:     444,
}

var MultiValidatorVote = primitives.MultiValidatorVote{
	Data:                  VoteData,
	Signature:             *sig,
	ParticipationBitfield: []byte{1, 2, 3, 4},
}
