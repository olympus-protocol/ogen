package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_BlockSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)

	blockheader := new(primitives.BlockHeader)
	sig := [96]byte{}
	rsig := [96]byte{}

	f.Fuzz(blockheader)
	f.Fuzz(&sig)
	f.Fuzz(&rsig)

	f.NumElements(32, 32)
	var deposits []*primitives.Deposit
	var exits []*primitives.Exit
	f.Fuzz(&deposits)
	f.Fuzz(&exits)

	f.NumElements(1000, 1000)
	var txs []*primitives.Tx
	f.Fuzz(&txs)

	f.NumElements(20, 20)
	var randaoSlash []*primitives.RANDAOSlashing
	f.Fuzz(&randaoSlash)

	f.NumElements(2, 2)
	var proposerSlash []*primitives.ProposerSlashing
	f.Fuzz(&proposerSlash)

	f.NumElements(128, 128)
	var governanceVotes []*primitives.GovernanceVote
	f.Fuzz(&governanceVotes)

	v := primitives.Block{
		Header:            blockheader,
		Votes:             nil,
		Txs:               txs,
		Deposits:          deposits,
		Exits:             exits,
		VoteSlashings:     nil,
		RANDAOSlashings:   randaoSlash,
		ProposerSlashings: proposerSlash,
		GovernanceVotes:   governanceVotes,
		Signature:         sig,
		RandaoSignature:   rsig,
	}

	ser, err := v.Marshal()
	assert.NoError(t, err)
	var desc primitives.Block
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}
