package primitives

import (
	"time"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var genesisHash = chainhash.Hash([chainhash.HashSize]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
})

// GetGenesisBlock gets the genesis block for a certain chain parameters.
func GetGenesisBlock(params params.ChainParams) Block {
	return Block{
		Header: BlockHeader{
			Version:        1,
			PrevBlockHash:  chainhash.Hash{},
			TxMerkleRoot:   chainhash.Hash{},
			VoteMerkleRoot: chainhash.Hash{},
			Timestamp:      uint64(time.Unix(0x0, 0).Unix()),
		},
		Txs: []Tx{},
	}
}
