package miner

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/wallet"
)

type Config struct {
	Log *logger.Logger
	// This miner key is only used to generate the first blocks before the workers network is up.
	MinerKey string
}

type Miner struct {
	log        *logger.Logger
	config     Config
	params     params.ChainParams
	chain      *chain.Blockchain
	walletsMan *wallet.WalletMan
	peersMan   *peers.PeerMan
	minerKey   *bls.SecretKey
	mineActive bool
}

func (m *Miner) Start() error {
	m.log.Info("Starting Miner instance")
	go m.MinerRoutine()
	return nil
}

func (m *Miner) Stop() {
	m.log.Info("Starting Miner instance")
}

func (m *Miner) MinerRoutine() {
	// check:
	// 	time.Sleep(time.Second * 5)
	// 	if !m.mineActive || !m.chain.State().IsSync() || m.peersMan.GetPeersCount() <= 0 {
	// 		goto check
	// 	}
	// 	m.log.Tracef("starting miner routine")
	// 	for {
	// 		block, err := m.createNewBlock()
	// 		if err != nil {
	// 			break
	// 		}
	// 		blockHash := block.Header.Hash()
	// 		m.log.Infof("created new block hash: %v txs: %v", blockHash, len(block.Txs))
	// 		m.peersMan.RelayBlockMsg(&p2p.MsgBlock{Block: *block})
	// 		break
	// 	}
	// 	time.Sleep(time.Second * 10)
	// 	goto check
}

// func (m *Miner) createNewBlock() (*primitives.Block, error) {
// state := m.chain.State().View.Tip()
// blockHeader := primitives.BlockHeader{
// 	Version:       1,
// 	PrevBlockHash: state.Hash,
// 	MerkleRoot:    chainhash.Hash{},
// 	Timestamp:     time.Now(),
// }
// blockHash := blockHeader.Hash()
// blockMsg := &primitives.Block{
// 	Header: blockHeader,
// 	Txs:    []primitives.Tx{},
// }
// sig, err := bls.Sign(m.minerKey, blockHash.CloneBytes())
// if err != nil {
// 	return nil, err
// }
// blockMsg.Signature = sig.Serialize()
// return blockMsg, nil
// }

func NewMiner(config Config, params params.ChainParams, chain *chain.Blockchain, walletsMan *wallet.WalletMan, peersMan *peers.PeerMan) (miner *Miner, err error) {
	var blsPrivKey bls.SecretKey
	var mineActive bool
	if config.MinerKey != "" {
		mineActive = true
		blsPrivKey, err = bls.NewSecretFromBech32(config.MinerKey, params.AddressPrefixes, false)
		if err != nil {
			return nil, err
		}
	} else {
		mineActive = false
	}
	miner = &Miner{
		log:        config.Log,
		config:     config,
		params:     params,
		chain:      chain,
		walletsMan: walletsMan,
		peersMan:   peersMan,
		minerKey:   &blsPrivKey,
		mineActive: mineActive,
	}
	return miner, nil
}
