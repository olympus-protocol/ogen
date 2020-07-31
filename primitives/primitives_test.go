package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/primitives"
)

func Test_BlockHeaderSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.BlockHeader
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.BlockHeader
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_BlockSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.Block
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Block
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_DepositSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.Deposit
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Deposit
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_ExitSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.Exit
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Exit
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_EpochReceiptSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.EpochReceipt
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.EpochReceipt
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_CommunityVoteDataSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.CommunityVoteData
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.CommunityVoteData
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_GovernanceVoteSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.GovernanceVote
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.GovernanceVote
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_VoteSlashingSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.VoteSlashing
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.VoteSlashing
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
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

func Test_ValidatorSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.Validator
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Validator
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_AcceptedVoteInfoSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.AcceptedVoteInfo
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.AcceptedVoteInfo
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_VoteDataSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.VoteData
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.VoteData
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_SingleValidatorVoteSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.SingleValidatorVote
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.SingleValidatorVote
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MultiValidatorVoteSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.MultiValidatorVote
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.MultiValidatorVote
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_CoinStateSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.CoinsState
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.CoinsState
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_GovernanceSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.Governance
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Governance
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_StateSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.State
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.State
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}
