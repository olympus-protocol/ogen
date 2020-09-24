package server

import (
	"context"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"net/http"

	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainrpc"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/proposer"
	"github.com/olympus-protocol/ogen/internal/wallet"
	"github.com/olympus-protocol/ogen/pkg/params"
)

type GlobalConfig struct {
	DataFolder string

	NetworkName string
	Port        string

	InitConfig state.InitializationParameters

	RPCProxy     bool
	RPCProxyPort string
	RPCProxyAddr string
	RPCPort      string
	RPCWallet    bool
	RPCAuthToken string

	Debug   bool
	LogFile bool
	Pprof   bool
}
type Mempools struct {
	Votes   mempool.VoteMempool
	Coins   mempool.CoinsMempool
	Actions mempool.ActionMempool
}

type Server interface {
	HostNode() hostnode.HostNode
	Proposer() proposer.Proposer
	Chain() chain.Blockchain
	Start()
	Stop() error
}

// Server is the main struct that contains ogen services
type server struct {
	log    logger.Logger
	config *GlobalConfig
	params params.ChainParams

	ch   chain.Blockchain
	hn   hostnode.HostNode
	rpc  chainrpc.RPCServer
	prop proposer.Proposer

	pools Mempools
}

var _ Server = &server{}

func (s *server) HostNode() hostnode.HostNode {
	return s.hn
}

func (s *server) Proposer() proposer.Proposer {
	return s.prop
}

func (s *server) Chain() chain.Blockchain {
	return s.ch
}

// Start starts running the multiple ogen services.
func (s *server) Start() {
	if s.config.Pprof {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}
	err := s.ch.Start()
	if err != nil {
		s.log.Fatal("unable to start chain instance")
	}
	err = s.hn.Start()
	if err != nil {
		s.log.Fatal("unable to start host node")
	}
	go func() {
		err := s.rpc.Start()
		if err != nil {
			s.log.Fatal("unable to start rpc server")
		}
	}()
	err = s.prop.Start()
	if err != nil {
		s.log.Fatal("unable to start proposer")
	}
}

// Stop closes the ogen services.
func (s *server) Stop() error {
	s.ch.Stop()
	s.rpc.Stop()
	s.hn.Stop()
	return nil
}

// NewServer creates a server instance and initializes the ogen services.
func NewServer(ctx context.Context, configParams *GlobalConfig, logger logger.Logger, currParams params.ChainParams, db blockdb.Database, ip state.InitializationParameters) (Server, error) {

	logger.Tracef("Loading network parameters for %v", currParams.Name)

	logger.Tracef("Initializing bls module with params for %v", currParams.Name)

	bls.Initialize(currParams)

	ch, err := chain.NewBlockchain(loadChainConfig(configParams, logger), currParams, db, ip)
	if err != nil {
		return nil, err
	}

	hn, err := hostnode.NewHostNode(ctx, loadPeersManConfig(configParams, logger), ch, currParams.NetMagic, false)
	if err != nil {
		return nil, err
	}

	lastActionManager, err := actionmanager.NewLastActionManager(ctx, hn, logger, ch, &currParams)
	if err != nil {
		return nil, err
	}

	coinsMempool, err := mempool.NewCoinsMempool(ctx, logger, ch, hn, &currParams)
	if err != nil {
		return nil, err
	}

	voteMempool, err := mempool.NewVoteMempool(ctx, logger, &currParams, ch, hn, lastActionManager)
	if err != nil {
		return nil, err
	}

	actionsMempool, err := mempool.NewActionMempool(ctx, logger, &currParams, ch, hn)
	if err != nil {
		return nil, err
	}

	voteMempool.Notify(actionsMempool)

	w, err := wallet.NewWallet(ctx, logger, configParams.DataFolder, &currParams, ch, hn, coinsMempool, actionsMempool)
	if err != nil {
		return nil, err
	}

	ks := keystore.NewKeystore(configParams.DataFolder, logger)

	prop, err := proposer.NewProposer(logger, &currParams, ch, hn, voteMempool, coinsMempool, actionsMempool, lastActionManager, ks)
	if err != nil {
		return nil, err
	}

	rpc, err := chainrpc.NewRPCServer(loadRPCConfig(configParams, logger), ch, hn, w, &currParams, prop)
	if err != nil {
		return nil, err
	}

	s := &server{
		config: configParams,
		log:    logger,

		ch:   ch,
		hn:   hn,
		rpc:  rpc,
		prop: prop,
		pools: Mempools{
			Votes:   voteMempool,
			Coins:   coinsMempool,
			Actions: actionsMempool,
		},
	}
	return s, nil
}

func loadChainConfig(config *GlobalConfig, logger logger.Logger) chain.Config {
	cfg := chain.Config{
		Log:     logger,
		Datadir: config.DataFolder,
	}
	return cfg
}

func loadPeersManConfig(config *GlobalConfig, logger logger.Logger) hostnode.Config {
	cfg := hostnode.Config{
		Log:  logger,
		Port: config.Port,
		Path: config.DataFolder,
	}
	return cfg
}

func loadRPCConfig(config *GlobalConfig, logger logger.Logger) chainrpc.Config {
	return chainrpc.Config{
		DataDir:      config.DataFolder,
		Log:          logger,
		RPCWallet:    config.RPCWallet,
		RPCProxy:     config.RPCProxy,
		RPCProxyPort: config.RPCProxyPort,
		RPCProxyAddr: config.RPCProxyAddr,
		RPCPort:      config.RPCPort,
		Network:      "tcp",
	}
}
