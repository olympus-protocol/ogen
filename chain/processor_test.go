package chain_test

import (
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/db/blockdb/mock"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"os"
	"testing"
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

	genesis, _ := b.State().View.Tip()
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

	genesis, _ := b.State().View.Tip()
	if genesis.Height != 0 {
		t.Fatal("expected genesis height to be 0")
	}

	if genesis.Parent != nil {
		t.Fatal("expected genesis parent to be nil")
	}

	//newBlockHeader := primitives.BlockHeader{
	//	Version:       0,
	//	Nonce:         0,
	//	MerkleRoot:    chainhash.Hash{},
	//	PrevBlockHash: genesis.Hash,
	//	Timestamp:     time.Time{},
	//}
	//
	//msgHash := newBlockHeader.Hash()
	//secretKey, _ := bls.RandSecretKey(rand.Reader)

	//sig, err := bls.Sign(secretKey, msgHash[:])
	//if err != nil {
	//	t.Fatal(err)
	//}

	//err = b.(&primitives.Block{
	//	Header: newBlockHeader,
	//	Txs:       nil,
	//	PubKey:    secretKey.DerivePublicKey().Serialize(),
	//	Signature: sig.Serialize(),
	//})
	//if err != nil {
	//	t.Fatal(err)
	//}
}

