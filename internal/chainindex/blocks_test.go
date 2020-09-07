package chainindex_test

import (
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBlockIndex_Instance(t *testing.T) {
	genblock := primitives.GetGenesisBlock()
	genesisHash := genblock.Hash()
	block2 := &primitives.Block{
		Header: &primitives.BlockHeader{
			Version:                    0,
			Nonce:                      0,
			TxMerkleRoot:               chainhash.Hash{},
			VoteMerkleRoot:             chainhash.Hash{},
			DepositMerkleRoot:          chainhash.Hash{},
			ExitMerkleRoot:             chainhash.Hash{},
			VoteSlashingMerkleRoot:     chainhash.Hash{},
			RANDAOSlashingMerkleRoot:   chainhash.Hash{},
			ProposerSlashingMerkleRoot: chainhash.Hash{},
			GovernanceVotesMerkleRoot:  chainhash.Hash{},
			PrevBlockHash:              genesisHash,
			Timestamp:                  uint64(time.Now().Unix()),
			Slot:                       1,
			StateRoot:                  chainhash.Hash{},
			FeeAddress:                 [20]byte{},
		},
		Signature:       [96]byte{},
		RandaoSignature: [96]byte{},
	}
	b2hash := block2.Hash()
	cIndex, err := chainindex.InitBlocksIndex(genblock)
	assert.NoError(t, err)
	assert.NotNil(t, cIndex)

	bRow, err := cIndex.Add(*block2)
	assert.NoError(t, err)
	assert.NotNil(t, bRow)

	bRow2, found := cIndex.Get(bRow.Hash)
	assert.Equal(t, true, found)
	assert.Equal(t, bRow.Hash, bRow2.Hash)

	have := cIndex.Have(b2hash)
	assert.Equal(t, true, have)

	bNode := bRow.ToBlockNodeDisk()

	bRow3, err := cIndex.LoadBlockNode(bNode)
	assert.NoError(t, err)
	assert.NotNil(t, bRow3)

	bRow3.AddChild(bRow)
	bRow3Child := bRow3.Children()
	assert.Equal(t, 1, len(bRow3Child))
	assert.Equal(t, bRow.Hash, bRow3Child[0].Hash)
}
