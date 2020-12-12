package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlockHeader(t *testing.T) {
	v := testdata.FuzzBlockHeader(10)
	for _, c := range v {
		ser, err := c.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.BlockHeader)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)
	}

	d := primitives.BlockHeader{
		Version:                    2,
		Nonce:                      10,
		TxMerkleRoot:               [32]byte{1, 2, 3},
		TxMultiMerkleRoot:          [32]byte{1, 2, 3},
		VoteMerkleRoot:             [32]byte{1, 2, 3},
		DepositMerkleRoot:          [32]byte{1, 2, 3},
		ExitMerkleRoot:             [32]byte{1, 2, 3},
		VoteSlashingMerkleRoot:     [32]byte{1, 2, 3},
		RANDAOSlashingMerkleRoot:   [32]byte{1, 2, 3},
		ProposerSlashingMerkleRoot: [32]byte{1, 2, 3},
		GovernanceVotesMerkleRoot:  [32]byte{1, 2, 3},
		PrevBlockHash:              [32]byte{1, 2, 3},
		Timestamp:                  500,
		Slot:                       14,
		StateRoot:                  [32]byte{1, 2, 3},
		FeeAddress:                 [20]byte{1, 2, 3},
	}

	assert.Equal(t, "d3f5389e43b628260ad8faa45a3f28d90f5d067884989161c0a0c9bdf5e6f80b", d.Hash().String())
}
