package server

import (
	"context"
	"log"
	"path"

	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chainrpc"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/mempool"
	"github.com/olympus-protocol/ogen/miner"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/logger"
	"github.com/olympus-protocol/ogen/wallet"
)

type Server struct {
	log    *logger.Logger
	config *config.Config
	params params.ChainParams

	Chain    *chain.Blockchain
	HostNode *peers.HostNode
	Wallet   *wallet.Wallet
	Miner    *miner.Miner
	RPC      *chainrpc.RPCServices
	Gui      bool
}

func (s *Server) Start() {
	err := s.Wallet.Start()
	if err != nil {
		log.Fatalf("unable to start wallet manager: %s", err)
	}
	err = s.Chain.Start()
	if err != nil {
		log.Fatalln("unable to start chain instance")
	}
	err = s.HostNode.Start()
	if err != nil {
		log.Fatalln("unable to start host node")
	}
	chainrpc.ServeRPC(s.RPC, loadRPCConfig(s.config, s.log))
	if s.Miner != nil {
		err = s.Miner.Start()
		if err != nil {
			log.Fatalln("unable to start miner thread")
		}
	}
}

func (s *Server) Stop() error {
	s.Chain.Stop()
	// s.HostNode.Stop()
	err := s.Wallet.Stop()
	if err != nil {
		return err
	}
	if s.Miner != nil {
		s.Miner.Stop()
	}
	return nil
}

func NewServer(ctx context.Context, configParams *config.Config, logger *logger.Logger, currParams params.ChainParams, db *bdb.BlockDB, gui bool, ip primitives.InitializationParameters) (*Server, error) {
	logger.Tracef("loading network parameters for '%v'", currParams.Name)
	ch, err := chain.NewBlockchain(loadChainConfig(configParams, logger), currParams, db, ip)
	if err != nil {
		return nil, err
	}
	walletConf := loadWalletsManConfig(configParams, logger)
	walletDB, err := badger.Open(badger.DefaultOptions(walletConf.Path).WithLogger(nil))
	if err != nil {
		return nil, err
	}
	hostnode, err := peers.NewHostNode(ctx, loadPeersManConfig(configParams, logger), ch, walletDB)
	if err != nil {
		return nil, err
	}
	coinsMempool, err := mempool.NewCoinsMempool(ctx, logger, ch, hostnode)
	if err != nil {
		return nil, err
	}
	voteMempool, err := mempool.NewVoteMempool(ctx, logger, &currParams, ch, hostnode)
	if err != nil {
		return nil, err
	}
	actionsMempool, err := mempool.NewActionMempool(ctx, logger, &currParams, ch, hostnode)
	if err != nil {
		return nil, err
	}
	w, err := wallet.NewWallet(ctx, walletConf, currParams, ch, hostnode, walletDB, coinsMempool, actionsMempool)
	if err != nil {
		return nil, err
	}
	rpc := &chainrpc.RPCServices{
		Wallet: chainrpc.NewRPCWallet(w, ch),
		Chain:  chainrpc.NewRPCChain(ch),
	}

	var min *miner.Miner
	if configParams.MiningEnabled {
		min, err = miner.NewMiner(loadMinerConfig(configParams, logger), currParams, ch, w, hostnode, voteMempool, coinsMempool, actionsMempool)
		if err != nil {
			return nil, err
		}
	}
	s := &Server{
		config: configParams,
		log:    logger,

		Chain:    ch,
		HostNode: hostnode,
		Wallet:   w,
		Miner:    min,
		Gui:      gui,
		RPC:      rpc,
	}
	return s, nil
}

func loadChainConfig(config *config.Config, logger *logger.Logger) chain.Config {
	cfg := chain.Config{
		Log: logger,
	}
	return cfg
}

func loadMinerConfig(config *config.Config, logger *logger.Logger) miner.Config {
	cfg := miner.Config{
		Log: logger,
	}
	return cfg
}

func loadPeersManConfig(config *config.Config, logger *logger.Logger) peers.Config {
	cfg := peers.Config{
		Log:      logger,
		Listen:   config.Listen,
		AddNodes: config.AddNodes,
		Port:     config.Port,
		MaxPeers: config.MaxPeers,
		Path:     config.DataFolder,
	}
	return cfg
}

func loadWalletsManConfig(config *config.Config, logger *logger.Logger) wallet.Config {
	cfg := wallet.Config{
		Log:  logger,
		Path: path.Join(config.DataFolder, "wallet"),
	}
	return cfg
}

func loadRPCConfig(config *config.Config, logger *logger.Logger) chainrpc.Config {
	return chainrpc.Config{
		Log:     logger,
		Address: config.RPCAddress,
		Network: "tcp",
	}
}
