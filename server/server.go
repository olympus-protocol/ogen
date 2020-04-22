package server

import (
	"log"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/explorer"
	"github.com/olympus-protocol/ogen/gov"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/mempool"
	"github.com/olympus-protocol/ogen/miner"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/users"
	"github.com/olympus-protocol/ogen/wallet"
	"github.com/olympus-protocol/ogen/workers"
)

type Server struct {
	log    *logger.Logger
	config *config.Config
	params params.ChainParams

	Chain     *chain.Blockchain
	PeerMan   *peers.PeerMan
	WalletMan *wallet.WalletMan
	Miner     *miner.Miner
	Mempool   *mempool.Mempool
	GovMan    *gov.GovMan
	WorkerMan *workers.WorkerMan
	UsersMan  *users.UserMan
	Gui       bool
}

func (s *Server) Start() {
	if s.config.Wallet {
		err := s.WalletMan.Start()
		if err != nil {
			log.Fatalln("unable to start wallet manager")
		}
	}
	err := s.Chain.Start()
	if err != nil {
		log.Fatalln("unable to start chain instance")
	}
	err = s.PeerMan.Start()
	if err != nil {
		log.Fatalln("unable to start peer manager")
	}
	err = s.Miner.Start()
	if err != nil {
		log.Fatalln("unable to start miner thread")
	}
	switch s.config.Mode {
	case "api":
		err := explorer.LoadApi(s.config, s.Chain, s.PeerMan)
		if err != nil {
			log.Fatal("unable to start api")
		}
	}
}

func (s *Server) Stop() error {
	s.Chain.Stop()
	s.PeerMan.Stop()
	if s.config.Wallet {
		err := s.WalletMan.Stop()
		if err != nil {
			return err
		}
	}
	s.Miner.Stop()
	return nil
}

func NewServer(configParams *config.Config, logger *logger.Logger, currParams params.ChainParams, db *blockdb.BlockDB, gui bool, ip primitives.InitializationParameters, keys miner.Keystore) (*Server, error) {
	logger.Tracef("loading network parameters for '%v'", params.NetworkNames[configParams.NetworkName])
	walletsMan, err := wallet.NewWalletMan(loadWalletsManConfig(configParams, logger, gui), currParams)
	if err != nil {
		return nil, err
	}
	ch, err := chain.NewBlockchain(loadChainConfig(configParams, logger), currParams, db, ip)
	if err != nil {
		return nil, err
	}
	peersMan, err := peers.NewPeersMan(loadPeersManConfig(configParams, logger), currParams, ch)
	if err != nil {
		return nil, err
	}
	min, err := miner.NewMiner(loadMinerConfig(configParams, logger), currParams, ch, walletsMan, peersMan, keys)
	if err != nil {
		return nil, err
	}
	txPool := mempool.InitMempool(loadMempoolConfig(configParams, logger), currParams)
	workersMan := workers.NewWorkersMan(loadWorkersConfig(configParams, logger), currParams)
	govMan := gov.NewGovMan(loadGovConfig(configParams, logger), currParams)
	usersMan := users.NewUsersMan(loadUsersConfig(configParams, logger), currParams)
	s := &Server{
		config: configParams,
		log:    logger,

		Chain:     ch,
		PeerMan:   peersMan,
		WalletMan: walletsMan,
		Miner:     min,
		Mempool:   txPool,
		WorkerMan: workersMan,
		GovMan:    govMan,
		UsersMan:  usersMan,
		Gui:       gui,
	}
	return s, nil
}

func loadGovConfig(config *config.Config, logger *logger.Logger) gov.Config {
	cfg := gov.Config{
		Log: logger,
	}
	return cfg
}

func loadUsersConfig(config *config.Config, logger *logger.Logger) users.Config {
	cfg := users.Config{
		Log: logger,
	}
	return cfg
}

func loadWorkersConfig(config *config.Config, logger *logger.Logger) workers.Config {
	cfg := workers.Config{
		Log: logger,
	}
	return cfg
}

func loadMempoolConfig(config *config.Config, logger *logger.Logger) mempool.Config {
	cfg := mempool.Config{
		Log: logger,
	}
	return cfg
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
		Log:          logger,
		Listen:       config.Listen,
		ConnectNodes: config.ConnectNodes,
		Port:         config.Port,
		MaxPeers:     config.MaxPeers,
		Path:         config.DataFolder,
	}
	return cfg
}

func loadWalletsManConfig(config *config.Config, logger *logger.Logger, gui bool) wallet.Config {
	cfg := wallet.Config{
		Log:      logger,
		Path:     config.DataFolder,
		Enabled:  config.Wallet,
		Gui:      gui,
	}
	return cfg
}
