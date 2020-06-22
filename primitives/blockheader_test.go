package primitives_test

import (
	"testing"

	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

var blockHeader primitives.BlockHeader

func init() {
	blockHeader = primitives.BlockHeader{
		Version:                    1,
		TxMerkleRoot:               chainhash.Hash{},
		VoteMerkleRoot:             chainhash.Hash{},
		DepositMerkleRoot:          chainhash.Hash{},
		ExitMerkleRoot:             chainhash.Hash{},
		VoteSlashingMerkleRoot:     chainhash.Hash{},
		RANDAOSlashingMerkleRoot:   chainhash.Hash{},
		ProposerSlashingMerkleRoot: chainhash.Hash{},
		GovernanceVotesMerkleRoot:  chainhash.Hash{},
		PrevBlockHash:              chainhash.Hash{},
		Timestamp:                  1592795022,
		Slot:                       0,
		StateRoot:                  chainhash.Hash{},
		FeeAddress:                 [20]byte{},
	}
}

func Test_Serialize(t *testing.T) {
	ser, err := blockHeader.Marshal()
	if err != nil {
		t.Error(err)
	}
	header := new(primitives.BlockHeader)
	err = header.Unmarshal(ser)
	if err != nil {
		t.Error(err)
	}
	equal := ssz.DeepEqual(blockHeader, header)
	if !equal {
		t.Error("masrhal/unmashal failed for blockheader")
	}
}
