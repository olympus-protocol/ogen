package miner

import (
	"context"
	"fmt"
	"time"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/wallet"
)

// Config is a config for the miner.
type Config struct {
	Log *logger.Logger
}

// Keystore is an interface to access keys.
type Keystore interface {
	GetKeyForWorker(w *primitives.Worker) (*bls.SecretKey, bool)
}

// BasicKeystore is a basic key store.
type BasicKeystore struct {
	keys map[[48]byte]bls.SecretKey
}

// GetKeyForWorker gets the key for a certain worker.
func (b *BasicKeystore) GetKeyForWorker(w *primitives.Worker) (*bls.SecretKey, bool) {
	key, found := b.keys[w.PubKey]
	return &key, found
}

// NewBasicKeystore creates a key store from the following keys.
func NewBasicKeystore(keys []bls.SecretKey) *BasicKeystore {
	m := make(map[[48]byte]bls.SecretKey)
	for _, k := range keys {
		pub := k.DerivePublicKey().Serialize()
		m[pub] = k
	}

	return &BasicKeystore{
		keys: m,
	}
}

func (m *Miner) Start() error {
	m.log.Info("starting miner")

	go func() {
		t := time.NewTimer(0)

	outer:
		for {
			select {
			case <-t.C:
				// check if we're an attester for this slot
				tip := m.chain.State().Tip().Header
				tipHash := tip.Hash()

				newSlot := tip.Slot + 1

				state := m.chain.State().TipState()

				min, max := state.GetVoteCommittee(newSlot, &m.params)

				data := primitives.VoteData{
					Slot:      newSlot,
					FromEpoch: state.JustifiedEpoch,
					FromHash:  state.JustifiedEpochHash,
					ToEpoch:   state.EpochIndex,
					ToHash:    state.GetRecentBlockHash(state.EpochIndex*m.params.EpochLength, &m.params),
				}

				dataHash := data.Hash()

				for i := min; i <= max; i++ {
					validator := state.ValidatorRegistry[i]

					if k, found := m.keystore.GetKeyForWorker(&validator); found {
						m.log.Infof("signing vote from %d to %d with validator %d", data.FromEpoch, data.ToEpoch, i)

						sig, err := bls.Sign(k, dataHash[:])
						if err != nil {
							panic(err)
						}

						vote := primitives.SingleValidatorVote{
							Data:      data,
							Signature: *sig,
							Offset:    i - min,
						}

						m.mempool.add(&vote, max-min)
					}
				}

				slotIndex := newSlot % m.params.EpochLength
				proposerIndex := state.ProposerQueue[slotIndex]
				proposer := state.ValidatorRegistry[proposerIndex]
				if k, found := m.keystore.GetKeyForWorker(&proposer); found {
					votes := m.mempool.get(newSlot, &m.params)

					block := primitives.Block{
						Header: primitives.BlockHeader{
							Version:       0,
							Nonce:         0,
							MerkleRoot:    chainhash.Hash{},
							PrevBlockHash: tipHash,
							Timestamp:     time.Now(),
							Slot:          newSlot,
						},
						Votes: votes,
					}

					blockHash := block.Hash()
					randaoHash := chainhash.HashH([]byte(fmt.Sprintf("%d", newSlot)))

					blockSig, err := bls.Sign(k, blockHash[:])
					if err != nil {
						panic(err)
					}
					randaoSig, err := bls.Sign(k, randaoHash[:])
					if err != nil {
						panic(err)
					}

					block.Signature = blockSig.Serialize()
					block.RandaoSignature = randaoSig.Serialize()
					if err := m.chain.ProcessBlock(&block); err != nil {
						panic(err)
					}
				}

				t = time.NewTimer(time.Second * 1)
			case <-m.context.Done():
				m.log.Info("stopping miner")
				break outer
			}
		}
	}()
	return nil
}

// Miner manages mining for the blockchain.
type Miner struct {
	log        *logger.Logger
	config     Config
	params     params.ChainParams
	chain      *chain.Blockchain
	walletsMan *wallet.WalletMan
	peersMan   *peers.PeerMan
	mineActive bool
	keystore   Keystore
	context    context.Context
	Stop       context.CancelFunc

	mempool *Mempool
}

// NewMiner creates a new miner from the parameters.
func NewMiner(config Config, params params.ChainParams, chain *chain.Blockchain, walletsMan *wallet.WalletMan, peersMan *peers.PeerMan, keys Keystore) (miner *Miner, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	miner = &Miner{
		log:        config.Log,
		config:     config,
		params:     params,
		chain:      chain,
		walletsMan: walletsMan,
		peersMan:   peersMan,
		mineActive: true,
		keystore:   keys,
		context:    ctx,
		Stop:       cancel,
		mempool:    NewMempool(),
	}
	return miner, nil
}
