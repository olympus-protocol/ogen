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

}

func Test_ExitSerialize(t *testing.T) {

}

func Test_EpochReceiptSerialize(t *testing.T) {

}

func Test_CommunityVoteDataSerialize(t *testing.T) {

}

func Test_GovernanceVoteSerialize(t *testing.T) {

}

func Test_VoteSlashingSerialize(t *testing.T) {

}

func Test_RANDAOSlashingSerialize(t *testing.T) {

}

func Test_ProposerSlashingSerialize(t *testing.T) {

}

func Test_TxLocatorSerialize(t *testing.T) {

}

func Test_TransferSinglePayloadSerialize(t *testing.T) {

}

func Test_TransferMultiPayloadSerialize(t *testing.T) {

}

func Test_TxSerialize(t *testing.T) {

}

func Test_ValidatorSerialize(t *testing.T) {

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
