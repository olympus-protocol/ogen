package serialization_test

import (
	"testing"

	"github.com/olympus-protocol/ogen/primitives"
	testdata "github.com/olympus-protocol/ogen/test/data"
	"github.com/prysmaticlabs/go-ssz"
)

func Test_BlockHeaderSerialize(t *testing.T) {
	ser, err := testdata.BlockHeader.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.BlockHeader
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.BlockHeader, desc)
	if !equal {
		t.Fatal("error: serialize BlockHeader")
	}
}

func Test_BlockSerialize(t *testing.T) {
	ser, err := testdata.Block.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.Block
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.Block, desc)
	if !equal {
		t.Fatal("error: serialize Block")
	}
}

func Test_DepositSerialize(t *testing.T) {
	ser, err := testdata.Deposit.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.Deposit
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.Deposit, desc)
	if !equal {
		t.Fatal("error: serialize Deposit")
	}
}

func Test_ExitSerialize(t *testing.T) {
	ser, err := testdata.Exit.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.Exit
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.Exit, desc)
	if !equal {
		t.Fatal("error: serialize Exit")
	}
}

func Test_EpochReceiptSerialize(t *testing.T) {
	ser, err := testdata.EpochReceipt.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.EpochReceipt
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.EpochReceipt, desc)
	if !equal {
		t.Fatal("error: serialize Exit")
	}
}

func Test_CommunityVoteDataSerialize(t *testing.T) {
	ser, err := testdata.CommunityVoteData.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.CommunityVoteData
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.CommunityVoteData, desc)
	if !equal {
		t.Fatal("error: serialize CommunityVoteData")
	}
}

func Test_GovernanceVoteSerialize(t *testing.T) {
	ser, err := testdata.GovernanceVote.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.GovernanceVote
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.GovernanceVote, desc)
	if !equal {
		t.Fatal("error: serialize GovernanceVote")
	}
}

func Test_VoteSlashingSerialize(t *testing.T) {
	ser, err := testdata.VoteSlashing.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.VoteSlashing
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.VoteSlashing, desc)
	if !equal {
		t.Fatal("error: serialize VoteSlashing")
	}
}

func Test_RANDAOSlashingSerialize(t *testing.T) {
	ser, err := testdata.RANDAOSlashing.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.RANDAOSlashing
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.RANDAOSlashing, desc)
	if !equal {
		t.Fatal("error: serialize RANDAOSlashing")
	}
}

func Test_ProposerSlashingSerialize(t *testing.T) {
	ser, err := testdata.ProposerSlashing.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.ProposerSlashing
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.ProposerSlashing, desc)
	if !equal {
		t.Fatal("error: serialize ProposerSlashing")
	}
}

func Test_TxLocatorSerialize(t *testing.T) {
	ser, err := testdata.TxLocator.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.TxLocator
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.TxLocator, desc)
	if !equal {
		t.Fatal("error: serialize TxLocator")
	}
}

func Test_TransferSinglePayloadSerialize(t *testing.T) {
	ser, err := testdata.TransferSinglePayload.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.TransferSinglePayload
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.TransferSinglePayload, desc)
	if !equal {
		t.Fatal("error: serialize TransferSinglePayload")
	}
}

func Test_TransferMultiPayloadSerialize(t *testing.T) {
	ser, err := testdata.TransferMultiPayload.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.TransferMultiPayload
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.TransferMultiPayload, desc)
	if !equal {
		t.Fatal("error: serialize TransferMultiPayload")
	}
}

func Test_TxSingleSerialize(t *testing.T) {
	ser, err := testdata.TxSingle.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.Tx
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.TxSingle, desc)
	if !equal {
		t.Fatal("error: serialize TxSingle")
	}
}

func Test_TxMultiSerialize(t *testing.T) {
	ser, err := testdata.TxMulti.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.Tx
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.TxMulti, desc)
	if !equal {
		t.Fatal("error: serialize TxMulti")
	}
}

func Test_ValidatorSerialize(t *testing.T) {
	ser, err := testdata.Validator.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.Validator
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.Validator, desc)
	if !equal {
		t.Fatal("error: serialize Validator")
	}
}

func Test_AcceptedVoteInfoSerialize(t *testing.T) {
	ser, err := testdata.AcceptedVoteInfo.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.AcceptedVoteInfo
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.AcceptedVoteInfo, desc)
	if !equal {
		t.Fatal("error: serialize AcceptedVoteInfo")
	}
}

func Test_VoteDataSerialize(t *testing.T) {
	ser, err := testdata.VoteData.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.VoteData
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.VoteData, desc)
	if !equal {
		t.Fatal("error: serialize VoteData")
	}
}

func Test_SingleValidatorVoteSerialize(t *testing.T) {
	ser, err := testdata.SingleValidatorVote.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.SingleValidatorVote
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.SingleValidatorVote, desc)
	if !equal {
		t.Fatal("error: serialize SingleValidatorVote")
	}
}

func Test_MultiValidatorVoteSerialize(t *testing.T) {
	ser, err := testdata.MultiValidatorVote.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var desc primitives.MultiValidatorVote
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Fatal(err)
	}
	equal := ssz.DeepEqual(testdata.MultiValidatorVote, desc)
	if !equal {
		t.Fatal("error: serialize MultiValidatorVote")
	}
}
