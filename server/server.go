package server

import (
	"context"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/peers/conflict"
	"net/http"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chainrpc"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/mempool"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/proposer"
	"github.com/olympus-protocol/ogen/utils/logger"
	"github.com/olympus-protocol/ogen/wallet"
)

// Server is the main struct that contains ogen services
type Server struct {
	log    *logger.Logger
	config *config.Config
	params params.ChainParams

	Chain    *chain.Blockchain
	HostNode *peers.HostNode
	Keystore *keystore.Keystore
	RPC      *chainrpc.RPCServer
	Proposer *proposer.Proposer
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
func NewServer(ctx context.Context, configParams *config.Config, logger *logger.Logger, currParams params.ChainParams, db *bdb.BlockDB, ip primitives.InitializationParameters) (*Server, error) {

	logger.Tracef("loading network parameters for %v", currParams.Name)

	logger.Tracef("initializing bls module with params for %v", currParams.Name)

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

	lastActionManager, err := conflict.NewLastActionManager(ctx, hostnode, logger, ch, &currParams)
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
	}
	return s, nil
}

func loadChainConfig(config *config.Config, logger *logger.Logger) chain.Config {
	cfg := chain.Config{
		Log:     logger,
		Datadir: config.DataFolder,
	}
	return cfg
}

func loadProposerConfig(config *config.Config, logger *logger.Logger) proposer.Config {
	cfg := proposer.Config{
		Datadir: config.DataFolder,
		Log:     logger,
	}
	return cfg
}

func loadPeersManConfig(config *config.Config, logger *logger.Logger) peers.Config {
	cfg := peers.Config{
		Log:      logger,
		AddNodes: config.AddNodes,
		Port:     config.Port,
		MaxPeers: config.MaxPeers,
		Path:     config.DataFolder,
	}
	return cfg
}

func loadRPCConfig(config *config.Config, logger *logger.Logger) chainrpc.Config {
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
