package state

import (
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type voterGroup struct {
	voters       map[uint64]struct{}
	totalBalance uint64
}

func (vg *voterGroup) add(id uint64, bal uint64) {
	if _, found := vg.voters[id]; found {
		return
	}

	vg.voters[id] = struct{}{}
	vg.totalBalance += bal
}

func (vg *voterGroup) addFromBitfield(registry []*primitives.Validator, field bitfield.Bitlist, validatorIndices []uint64) {
	for i, validatorIdx := range validatorIndices {
		if field.Get(uint(i)) {
			vg.add(validatorIdx, registry[validatorIdx].Balance)
		}
	}
}

func (vg *voterGroup) contains(id uint64) bool {
	_, found := vg.voters[id]
	return found
}

func newVoterGroup() voterGroup {
	return voterGroup{
		voters: make(map[uint64]struct{}),
	}
}
