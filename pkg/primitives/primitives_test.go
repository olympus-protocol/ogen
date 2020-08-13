package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/bitfield"

	"github.com/olympus-protocol/ogen/pkg/primitives"
)

func fuzzVoteSlashing(n int) []*primitives.VoteSlashing {
	var votes []*primitives.VoteSlashing
	for i := 0; i < n; i++ {
		v := &primitives.VoteSlashing{
			Vote1: fuzzMultiValidatorVote(1)[0],
			Vote2: fuzzMultiValidatorVote(1)[0],
		}
		votes = append(votes, v)
	}
	return votes
}

func fuzzMultiValidatorVote(n int) []*primitives.MultiValidatorVote {
	var votes []*primitives.MultiValidatorVote
	for i := 0; i < n; i++ {
		f := fuzz.New().NilChance(0)
		d := new(primitives.VoteData)
		var sig [96]byte
		f.Fuzz(d)
		f.Fuzz(&sig)
		v := &primitives.MultiValidatorVote{
			Data:                  d,
			Sig:                   sig,
			ParticipationBitfield: bitfield.NewBitlist(10),
		}
		votes = append(votes, v)
	}
	return votes
}

func fuzzAcceptedVoteInfo(n int) []*primitives.AcceptedVoteInfo {
	var avInfo []*primitives.AcceptedVoteInfo
	for i := 0; i < n; i++ {
		f := fuzz.New().NilChance(0)
		d := new(primitives.VoteData)
		var sig [96]byte
		f.Fuzz(d)
		f.Fuzz(&sig)
		v := &primitives.AcceptedVoteInfo{
			Data:                  d,
			ParticipationBitfield: bitfield.NewBitlist(10),
			Proposer:              0,
			InclusionDelay:        0,
		}
		avInfo = append(avInfo, v)
	}
	return avInfo
}
