package primitives_test

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestBlock(t *testing.T) {
	correct := testdata.FuzzBlock(2, true, true)
	for _, b := range correct {
		ser, err := b.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.Block)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, b, desc)
	}

	incorrect := testdata.FuzzBlock(2, false, true)
	for _, b := range incorrect {
		_, err := b.Marshal()
		assert.NotNil(t, err)

	}

	nilpointers := testdata.FuzzBlock(2, true, false)
	for _, b := range nilpointers {
		assert.NotPanics(t, func() {
			data, err := b.Marshal()
			assert.NoError(t, err)

			n := new(primitives.Block)
			err = n.Unmarshal(data)
			assert.NoError(t, err)

			assert.Equal(t, b, n)

			assert.Equal(t, uint64(0), n.Header.Slot)
		})

	}
}

func TestBlocksMerkle(t *testing.T) {
	// Serialized snappy compressed block
	blockRaw, err := os.ReadFile("./block_raw.dat")
	assert.NoError(t, err)

	assert.NoError(t, err)

	b := new(primitives.Block)
	err = b.Unmarshal(blockRaw)
	assert.NoError(t, err)

	assert.Equal(t, "c7261a02a4693b496e12fe7601ef919618d62fe4d5b5f5c8a394c81c3879d044", b.VotesMerkleRoot().String())
	assert.Equal(t, "3627179c10539597902e23bfdbc62b34a8ed7588d9043bb9d94a72b93f53e3f7", b.DepositMerkleRoot().String())
	assert.Equal(t, "28bc1ef7f81e8a002076a50e12636ff6c4f3ee968b3c44e0f4a264a671c96bfb", b.ExitMerkleRoot().String())
	assert.Equal(t, "78a4602fd8a59330ea99f8739c7525b39cfb33ecf4fbb1d61de8da39e4894816", b.PartialExitsMerkleRoot().String())
	assert.Equal(t, "f646c46922e0eec38c69a69eceed0605c2f1411ca816dae1b354832d743e8ff3", b.CoinProofsMerkleRoot().String())
	assert.Equal(t, "7d80ed97f4ac4f03e447cbc0ad693be2d8503895187f2e370d05f93a50b9b20d", b.ExecutionsMerkleRoot().String())
	assert.Equal(t, "aa7fe3c283f3a9443d777756b3d1d3736a4312bbf2317d68e6ebcdee9e2334e0", b.TxsMerkleRoot().String())
	assert.Equal(t, "a041e5f7633f037b26c58e090e08862fdbfd98375fa55312147faa15aa1ad3e0", b.VoteSlashingRoot().String())
	assert.Equal(t, "e2df0f1c2f51a57e01c85096c5953929f786943f48081bfa53f64202b1b23d37", b.ProposerSlashingsRoot().String())
	assert.Equal(t, "739b84029f9349a43a4b677e01a46f0e84a9347234904eae665007e3d63c34f4", b.RANDAOSlashingsRoot().String())
	assert.Equal(t, "5ca8e62f3407d108e6de53caf98c17ec7f7646a0e64ccd049cb1f79665f7ec6a", b.GovernanceVoteMerkleRoot().String())
	assert.Equal(t, "e21e014179d5295c527764c4e3f747af1636dc2ad5f14463c8fbe8d857e8e031", b.MultiSignatureTxsMerkleRoot().String())

	expectedTx := []string{"8408d61bc2dea7b8548ce514556640457fc54346c883afa35430dcb26dd43bb5", "b6e5501e7c81955342e94927fcc92e16ef11ce210e8d56f33942147e708ef4c5"}
	txs := b.GetTxs()

	assert.Equal(t, expectedTx[0], txs[0])
	assert.Equal(t, expectedTx[1], txs[1])
}
