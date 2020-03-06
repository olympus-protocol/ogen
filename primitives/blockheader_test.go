package primitives

import (
	"bytes"
	"github.com/go-test/deep"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"testing"
	"time"
)

var (
	merkleRootTest = chainhash.Hash([chainhash.HashSize]byte{
		0xfc, 0x4b, 0x8c, 0xb9, 0x03, 0xae, 0xd5, 0x4e,
		0x11, 0xe1, 0xae, 0x8a, 0x5b, 0x7a, 0xd0, 0x97,
		0xad, 0xe3, 0x49, 0x88, 0xa8, 0x45, 0x00, 0xad,
		0x2d, 0x80, 0xe4, 0xd1, 0xf5, 0xbc, 0xc9, 0x5d,
	})
	blockHeaderTest = BlockHeader{
		Version:       1,
		PrevBlockHash: chainhash.Hash{},
		MerkleRoot:    merkleRootTest,
		Timestamp:     time.Unix(0x5A3BB72B, 0),
	}
)

func TestBlockHeader_Serialize(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := blockHeaderTest.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}
	var blockHeader BlockHeader
	err = blockHeader.Deserialize(buf)
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(blockHeader, blockHeaderTest); diff != nil {
		t.Fatal(diff)
	}
}
