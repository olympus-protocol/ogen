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

		assert.Equal(t, primitives.BlockHeaderSize, len(ser))

		desc := new(primitives.BlockHeader)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)
	}

	d := primitives.BlockHeader{
		Version:                     2,
		PrevBlockHash:               [32]byte{1, 2, 3},
		Timestamp:                   500,
		Slot:                        14,
		FeeAddress:                  [20]byte{1, 2, 3},
		VoteMerkleRoot:              [32]byte{1, 2, 3},
		DepositMerkleRoot:           [32]byte{1, 2, 3},
		ExitMerkleRoot:              [32]byte{1, 2, 3},
		PartialExitMerkleRoot:       [32]byte{1, 2, 3},
		CoinProofsMerkleRoot:        [32]byte{1, 2, 3},
		ExecutionsMerkleRoot:        [32]byte{1, 2, 3},
		TxsMerkleRoot:               [32]byte{1, 2, 3},
		VoteSlashingMerkleRoot:      [32]byte{1, 2, 3},
		ProposerSlashingMerkleRoot:  [32]byte{1, 2, 3},
		RANDAOSlashingMerkleRoot:    [32]byte{1, 2, 3},
		GovernanceVotesMerkleRoot:   [32]byte{1, 2, 3},
		MultiSignatureTxsMerkleRoot: [32]byte{1, 2, 3},
	}

	assert.Equal(t, "6969a416b73c94c0525922c8b5581fdb209b56dc47a28dca4abe660405cc0d72", d.Hash().String())
}
