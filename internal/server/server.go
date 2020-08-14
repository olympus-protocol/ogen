package server

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"net/http"

	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainrpc"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/peers"
	"github.com/olympus-protocol/ogen/internal/proposer"
	"github.com/olympus-protocol/ogen/internal/wallet"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type GlobalConfig struct {
	DataFolder string

	NetworkName  string
	InitialNodes []peer.AddrInfo
	Port         string

	InitConfig primitives.InitializationParameters

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
	Votes   *mempool.VoteMempool
	Coins   *mempool.CoinsMempool
	Actions *mempool.ActionMempool
}

// Server is the main struct that contains ogen services
type Server struct {
	log    *logger.Logger
	config *GlobalConfig
	params params.ChainParams

	Chain    chain.Blockchain
	HostNode peers.HostNode
	RPC      *chainrpc.RPCServer
	Proposer proposer.Proposer

	Mempools Mempools
}

// Start starts running the multiple ogen services.
func (s *Server) Start() {
	if s.config.Pprof {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}
	err := s.Chain.Start()
	if err != nil {
		s.log.Fatal("unable to start chain instance")
	}
	err = s.HostNode.Start()
	if err != nil {
		s.log.Fatal("unable to start host node")
	}
	go func() {
		err := s.RPC.Start()
		if err != nil {
			s.log.Fatal("unable to start rpc server")
		}
	}()
}

// Stop closes the ogen services.
func (s *Server) Stop() error {
	s.Chain.Stop()
	s.RPC.Stop()
	return nil
}

// NewServer creates a server instance and initializes the ogen services.
func NewServer(ctx context.Context, configParams *GlobalConfig, logger *logger.Logger, currParams params.ChainParams, db blockdb.BlockDB, ip primitives.InitializationParameters) (*Server, error) {

	logger.Tracef("Loading network parameters for %v", currParams.Name)

	logger.Tracef("Initializing bls module with params for %v", currParams.Name)

	err := bls.Initialize(currParams)
	if err != nil {
		return nil, err
	}

	ch, err := chain.NewBlockchain(loadChainConfig(configParams, logger), currParams, db, ip)
	if err != nil {
		return nil, err
	}

	hostnode, err := peers.NewHostNode(ctx, loadPeersManConfig(configParams, logger), ch)
	if err != nil {
		return nil, err
	}

	lastActionManager, err := actionmanager.NewLastActionManager(ctx, hostnode, logger, ch, &currParams)
	if err != nil {
		return nil, err
	}

	coinsMempool, err := mempool.NewCoinsMempool(ctx, logger, ch, hostnode, &currParams)
	if err != nil {
		return nil, err
	}

	voteMempool, err := mempool.NewVoteMempool(ctx, logger, &currParams, ch, hostnode, lastActionManager)
	if err != nil {
		return nil, err
	}

	actionsMempool, err := mempool.NewActionMempool(ctx, logger, &currParams, ch, hostnode)
	if err != nil {
		return nil, err
	}

	voteMempool.Notify(actionsMempool)

	w, err := wallet.NewWallet(ctx, logger, configParams.DataFolder, &currParams, ch, hostnode, coinsMempool, actionsMempool)
	if err != nil {
		return nil, err
	}

	prop, err := proposer.NewProposer(loadProposerConfig(configParams, logger), currParams, ch, hostnode, voteMempool, coinsMempool, actionsMempool, lastActionManager)
	if err != nil {
		return nil, err
	}

	rpc, err := chainrpc.NewRPCServer(loadRPCConfig(configParams, logger), ch, hostnode, w, &currParams, prop)
	if err != nil {
		return nil, err
	}

	s := &Server{
		config: configParams,
		log:    logger,

		Chain:    ch,
		HostNode: hostnode,
		RPC:      rpc,
		Proposer: prop,
		Mempools: Mempools{
			Votes:   voteMempool,
			Coins:   coinsMempool,
			Actions: actionsMempool,
		},
	}
	return s, nil
}

func loadChainConfig(config *GlobalConfig, logger *logger.Logger) chain.Config {
	cfg := chain.Config{
		Log:     logger,
		Datadir: config.DataFolder,
	}
	return cfg
}

func loadProposerConfig(config *GlobalConfig, logger *logger.Logger) proposer.Config {
	cfg := proposer.Config{
		Datadir: config.DataFolder,
		Log:     logger,
	}
	return cfg
}

func loadPeersManConfig(config *GlobalConfig, logger *logger.Logger) peers.Config {
	cfg := peers.Config{
		Log:          logger,
		InitialNodes: config.InitialNodes,
		Port:         config.Port,
		Path:         config.DataFolder,
	}
	return cfg
}

func loadRPCConfig(config *GlobalConfig, logger *logger.Logger) chainrpc.Config {
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
