package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_VoteSlashingSerialize(t *testing.T) {
	//v := fuzzVoteSlashing(1)
	//ser, err := v[0].Marshal()
	//assert.NoError(t, err)
	//
	//desc := new(primitives.VoteSlashing)
	//err = desc.Unmarshal(ser)
	//assert.NoError(t, err)
	//
	//assert.Equal(t, v[0], desc)
}

func Test_RANDAOSlashingSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.RANDAOSlashing
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)
	var desc primitives.RANDAOSlashing
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_ProposerSlashingSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.ProposerSlashing
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.ProposerSlashing
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}
