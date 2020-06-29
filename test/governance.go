package testdata

import (
	"github.com/olympus-protocol/ogen/primitives"
)

var CommunityVoteData = primitives.CommunityVoteData{
	ReplacementCandidates: [][20]byte{
		{252, 94, 117, 132, 63, 93, 202, 26, 36, 23, 195, 26, 169, 95, 74, 147, 72, 184, 66, 20},
		{252, 94, 117, 132, 63, 93, 202, 26, 36, 23, 195, 26, 169, 95, 74, 147, 72, 184, 66, 20},
		{252, 94, 117, 132, 63, 93, 202, 26, 36, 23, 195, 26, 169, 95, 74, 147, 72, 184, 66, 20},
		{252, 94, 117, 132, 63, 93, 202, 26, 36, 23, 195, 26, 169, 95, 74, 147, 72, 184, 66, 20},
		{252, 94, 117, 132, 63, 93, 202, 26, 36, 23, 195, 26, 169, 95, 74, 147, 72, 184, 66, 20},
	},
}

var GovernanceVote = primitives.GovernanceVote{
	Type:          1,
	Data:          []byte{12, 12, 12, 13, 13, 13},
	FunctionalSig: funcSigByte,
	VoteEpoch:     100,
}
