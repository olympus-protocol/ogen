package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// fuzzAcceptedVoteInfo return a slice with n AcceptedVoteInfo structs.
// If correct is true will return correctly serializable structs
// If complete is true will return information with no nil pointers.
func fuzzAcceptedVoteInfo(n int, correct bool, complete bool) []*primitives.AcceptedVoteInfo {
	var v []*primitives.AcceptedVoteInfo
	for i := 0; i < n; i++ {
		i := new(primitives.AcceptedVoteInfo)
		f := fuzz.New().NilChance(0)
		f.Fuzz(i)
		i.ParticipationBitfield = bitfield.NewBitlist(6242)
		if !correct {
			i.ParticipationBitfield = bitfield.NewBitlist(50000)
		}
		if !complete {
			i.Data = nil
		}
		v = append(v, i)
	}
	return v
}
