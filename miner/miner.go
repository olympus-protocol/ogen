package miner

import (
	"context"
	"fmt"
	"time"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chain/index"
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
	GetKey(w *primitives.Worker) (*bls.SecretKey, bool)
}

// BasicKeystore is a basic key store.
type BasicKeystore struct {
	keys map[[48]byte]bls.SecretKey
}

// GetKey gets the key for a certain worker.
func (b *BasicKeystore) GetKey(w *primitives.Worker) (*bls.SecretKey, bool) {
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
	chain.Notify(miner)
	return miner, nil
}

// NewTip implements the BlockchainNotifee interface.
func (m *Miner) NewTip(row *index.BlockRow, block *primitives.Block) {
	m.mempool.remove(block)
}

func (m *Miner) getCurrentSlot() uint64 {
	slot := int64(time.Now().Sub(m.chain.GenesisTime()) / (time.Duration(m.params.SlotDuration) * time.Second))
	if slot < 0 {
		return 0
	}
	return uint64(slot)
}

// getNextSlotTime gets the next slot time.
func (m *Miner) getNextBlockTime(nextSlot uint64) time.Time {
	return m.chain.GenesisTime().Add(time.Duration(nextSlot*m.params.SlotDuration) * time.Second)
}

// getNextSlotTime gets the next slot time.
func (m *Miner) getNextVoteTime(nextSlot uint64) time.Time {
	return m.chain.GenesisTime().Add(time.Duration(nextSlot*m.params.SlotDuration) * time.Second).Add(-time.Second * time.Duration(m.params.SlotDuration) / 2)
}

// Start runs the miner.
func (m *Miner) Start() error {
	m.log.Info("starting miner")

	go func() {
		slotToPropose := m.getCurrentSlot() + 1
		slotToVote := slotToPropose
		blockTimer := time.NewTimer(time.Until(m.getNextBlockTime(slotToPropose)))
		voteTimer := time.NewTimer(time.Until(m.getNextVoteTime(slotToVote)))

	outer:
		for {
			select {
			case <-voteTimer.C:
				// check if we're an attester for this slot
				tip := m.chain.State().Tip()
				tipHash := tip.Hash

				m.log.Infof("sending votes for slot %d", slotToVote)

				s := m.chain.State()

				view, err := s.GetSubView(tipHash)
				if err != nil {
					panic(err)
				}

				state, err := s.GetStateForHashAtSlot(tipHash, slotToVote, &view, &m.params)
				if err != nil {
					panic(err)
				}

				min, max := state.GetVoteCommittee(slotToVote, &m.params)
				toEpoch := (slotToVote - 1) / m.params.EpochLength

				data := primitives.VoteData{
					Slot:      slotToVote,
					FromEpoch: state.JustifiedEpoch,
					FromHash:  state.JustifiedEpochHash,
					ToEpoch:   toEpoch,
					ToHash:    state.GetRecentBlockHash(toEpoch*m.params.EpochLength-1, &m.params),
				}

				dataHash := data.Hash()

				for i := min; i <= max; i++ {
					validator := state.ValidatorRegistry[i]

					if k, found := m.keystore.GetKey(&validator); found {
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
				slotToVote++
				voteTimer = time.NewTimer(time.Until(m.getNextVoteTime(slotToVote)))

			case <-blockTimer.C:
				// check if we're an attester for this slot
				tip := m.chain.State().Tip()
				tipHash := tip.Hash

				s := m.chain.State()

				m.log.Infof("proposing for slot %d", slotToPropose)

				view, err := s.GetSubView(tipHash)
				if err != nil {
					panic(err)
				}

				state, err := s.GetStateForHashAtSlot(tipHash, slotToPropose, &view, &m.params)
				if err != nil {
					panic(err)
				}

				slotIndex := (slotToPropose + m.params.EpochLength - 1) % m.params.EpochLength

				proposerIndex := state.ProposerQueue[slotIndex]
				proposer := state.ValidatorRegistry[proposerIndex]

				if k, found := m.keystore.GetKey(&proposer); found {
					votes := m.mempool.get(slotToPropose, &m.params)

					block := primitives.Block{
						Header: primitives.BlockHeader{
							Version:       0,
							Nonce:         0,
							MerkleRoot:    chainhash.Hash{},
							PrevBlockHash: tipHash,
							Timestamp:     time.Now(),
							Slot:          slotToPropose,
						},
						Votes: votes,
					}

					blockHash := block.Hash()
					randaoHash := chainhash.HashH([]byte(fmt.Sprintf("%d", slotToPropose)))

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

				slotToPropose++
				blockTimer = time.NewTimer(time.Until(m.getNextBlockTime(slotToPropose)))
			case <-m.context.Done():
				m.log.Info("stopping miner")
				break outer
			}
		}
	}()
	return nil
}
