package chain_test

import (
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/db/blockdb/mock"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"os"
	"testing"
	"time"
)

var log = logger.New(os.Stdout).Quiet()


func TestBlockchainTipGenesis(t *testing.T) {
	db := mock.NewMemoryDB()

	b, err := chain.NewBlockchain(chain.Config{
		Log: log,
	}, params.Mainnet, db)
	if err != nil {
		t.Fatal(err)
	}

	genesis := b.State().View.Tip()
	if genesis.Height != 0 {
		t.Fatal("expected genesis height to be 0")
	}

	if genesis.Parent != nil {
		t.Fatal("expected genesis parent to be nil")
	}
}

func TestBlockchainTipAddBlock(t *testing.T) {
	db := mock.NewMemoryDB()

	b, err := chain.NewBlockchain(chain.Config{
		Log: log,
	}, params.Mainnet, db)
	if err != nil {
		t.Fatal(err)
	}

	genesis := b.State().View.Tip()
	if genesis.Height != 0 {
		t.Fatal("expected genesis height to be 0")
	}

	if genesis.Parent != nil {
		t.Fatal("expected genesis parent to be nil")
	}

	err = b.ProcessBlock(&primitives.Block{
		Header: primitives.BlockHeader{
			Version:       0,
			Nonce:         0,
			MerkleRoot:    chainhash.Hash{},
			PrevBlockHash: genesis.Hash,
			Timestamp:     time.Time{},
		},
		Txs:       nil,
		PubKey:    [48]byte{},
		Signature: [96]byte{},
	})
	if err != nil {
		t.Fatal(err)
	}
}

