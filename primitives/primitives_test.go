package primitives_test

import (
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/stretchr/testify/assert"
)

func Test_BlockHeaderSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)

	v := new(primitives.BlockHeader)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.BlockHeader)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_BlockSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.Block)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.Block)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_DepositSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.Deposit)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.Deposit)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_ExitSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.Exit)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.Exit)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_EpochReceiptSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.EpochReceipt)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.EpochReceipt)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_CommunityVoteDataSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.CommunityVoteData)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.CommunityVoteData)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_GovernanceVoteSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.GovernanceVote)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.GovernanceVote)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_VoteSlashingSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.VoteSlashing)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.VoteSlashing)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_RANDAOSlashingSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.RANDAOSlashing)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.RANDAOSlashing)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_ProposerSlashingSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.ProposerSlashing)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.ProposerSlashing)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_TransferSinglePayloadSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.TransferSinglePayload)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.TransferSinglePayload)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_TransferMultiPayloadSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.TransferMultiPayload)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.TransferMultiPayload)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_TxSingleSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.Tx)

	p := new(primitives.TransferSinglePayload)

	f.Fuzz(p)

	v.AppendPayload(p)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.Tx)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_TxMultiSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.Tx)

	p := new(primitives.TransferMultiPayload)

	f.Fuzz(p)

	v.AppendPayload(p)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.Tx)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_ValidatorSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.Validator)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.Validator)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_AcceptedVoteInfoSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.AcceptedVoteInfo)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.AcceptedVoteInfo)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_VoteDataSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.VoteData)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.VoteData)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_SingleValidatorVoteSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.SingleValidatorVote)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.SingleValidatorVote)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MultiValidatorVoteSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.MultiValidatorVote)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.MultiValidatorVote)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_CoinStateSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.CoinsState)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.CoinsState)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_GovernanceSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.Governance)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.Governance)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_StateSerialize(t *testing.T) {

	f := fuzz.New().NilChance(0)

	v := new(primitives.State)

	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(primitives.State)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}
