package index_test

import (
	"bytes"
	"github.com/go-test/deep"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"testing"
	"time"
)

func TestSerializeDeserializeRow(t *testing.T) {
	header := primitives.BlockHeader{
		Version:       1,
		Nonce:         2,
		MerkleRoot:    chainhash.Hash{3},
		PrevBlockHash: chainhash.Hash{4},
		Timestamp:     time.Unix(5, 0),
	}
	blockRow := index.BlockRow{
		Header: header,
		Locator: blockdb.BlockLocation{},
		Height:  7,
		Parent:  nil,
		Hash: header.Hash(),
	}

	b := bytes.NewBuffer([]byte{})
	err := blockRow.Serialize(b)
	if err != nil {
		t.Fatal(err)
	}

	blockRowDeser := index.BlockRow{}
	err = blockRowDeser.Deserialize(b)
	if err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(blockRow, blockRowDeser); diff != nil {
		t.Fatal(diff)
	}
}

func TestSerializeDeserializeIndex(t *testing.T) {
	genesisHeader := primitives.BlockHeader{
		Version:       1,
		Nonce:         2,
		MerkleRoot:    chainhash.Hash{3},
		PrevBlockHash: chainhash.Hash{4},
		Timestamp:     time.Unix(5, 0),
	}
	blockIndex, err := index.InitBlocksIndex(genesisHeader, blockdb.BlockLocation{})
	if err != nil {
		t.Fatal(err)
	}

	genesisHash := genesisHeader.Hash()

	blockHeader := primitives.BlockHeader{
		Version:       1,
		Nonce:         2,
		MerkleRoot:    chainhash.Hash{3},
		PrevBlockHash: genesisHash,
		Timestamp:     time.Unix(5, 0),
	}

	_, err = blockIndex.Add(blockHeader, blockdb.BlockLocation{})
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBuffer([]byte{})
	err = blockIndex.Serialize(b)
	if err != nil {
		t.Fatal(err)
	}

	blockIndexDeser := &index.BlockIndex{}
	err = blockIndexDeser.Deserialize(b)
	if err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(blockIndex, blockIndexDeser); diff != nil {
		t.Fatal(diff)
	}

	blockHash := blockHeader.Hash()

	blockIndexHeader, _ := blockIndex.Get(blockHash)
	blockIndexDeserHeader, _ := blockIndexDeser.Get(blockHash)

	if diff := deep.Equal(blockIndexHeader, blockIndexDeserHeader); diff != nil {
		t.Fatal(diff)
	}
}
