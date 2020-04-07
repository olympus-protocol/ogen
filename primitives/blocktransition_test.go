package primitives_test

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/pkg/errors"
)

const numTestValidators = 128

func getTestInitializationParameters() (*primitives.InitializationParameters, []bls.SecretKey) {
	vals := make([]primitives.ValidatorInitialization, numTestValidators)
	keys := make([]bls.SecretKey, numTestValidators)
	for i := range vals {
		k, err := bls.RandSecretKey(rand.Reader)
		if err != nil {
			panic(err)
		}

		keys[i] = *k

		vals[i] = primitives.ValidatorInitialization{
			PubKey:       keys[i].DerivePublicKey().Serialize(),
			PayeeAddress: "",
		}
	}

	return &primitives.InitializationParameters{
		InitialValidators: vals,
	}, keys
}

func signedBlock(s *primitives.State, keys []bls.SecretKey, p *params.ChainParams, b primitives.Block) primitives.Block {
	blockHash := b.Hash()

	slotIndex := b.Header.Slot % p.EpochLength
	proposerIndex := s.ProposerQueue[slotIndex]

	sig, err := bls.Sign(&keys[proposerIndex], blockHash[:])
	if err != nil {
		panic(err)
	}

	randaoHash := chainhash.HashH([]byte(fmt.Sprintf("%d", b.Header.Slot)))
	randaoSig, err := bls.Sign(&keys[proposerIndex], randaoHash[:])
	if err != nil {
		panic(err)
	}

	b.Signature = sig.Serialize()
	b.RandaoSignature = randaoSig.Serialize()

	return b
}

var zeroHash = chainhash.Hash{}

func TestProcessBlock(t *testing.T) {
	genesisBlock := primitives.GetGenesisBlock(params.TestNet)
	genesisHash := genesisBlock.Hash()

	ip, keys := getTestInitializationParameters()

	p := &params.TestNet

	startingState := primitives.GetGenesisStateWithInitializationParameters(genesisHash, ip, &params.TestNet)

	blockTests := []struct {
		name  string
		block primitives.Block
		check func(b *primitives.Block, s *primitives.State) error
	}{
		{
			name: "Empty block",
			block: signedBlock(startingState, keys, p, primitives.Block{
				Header: primitives.BlockHeader{
					Version:   0,
					Nonce:     0,
					Timestamp: time.Time{},
					Slot:      0,
				},
				Votes: nil,
				Txs:   nil,
			}),
			check: func(b *primitives.Block, s *primitives.State) error {
				expectedRandao := startingState.RANDAO

				for i := range expectedRandao {
					expectedRandao[i] ^= b.RandaoSignature[i]
				}

				if !s.NextRANDAO.IsEqual(&expectedRandao) {
					return fmt.Errorf("expected NextRANDAO to be: %s, got: %s", expectedRandao, s.NextRANDAO)
				}
				return nil
			},
		},
	}

	for _, test := range blockTests {
		newState := startingState.Copy()
		err := newState.ProcessBlock(&test.block, &params.TestNet)

		if err != nil {
			t.Fatal(errors.Wrapf(err, "error running test %s", test.name))
		}

		if err := test.check(&test.block, &newState); err != nil {
			t.Fatal(errors.Wrapf(err, "error running test %s", test.name))
		}
	}
}
