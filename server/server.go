package server

import (
	"context"
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

type Server struct {
	log    *logger.Logger
	config *config.Config
	params params.ChainParams

	Chain    *chain.Blockchain
	HostNode *peers.HostNode
	Keystore *keystore.Keystore
	Proposer *proposer.Proposer
	RPC      *chainrpc.RPCServer
}

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
	if s.Proposer != nil {
		err = s.Proposer.Start()
		if err != nil {
			s.log.Fatal("unable to start proposer thread")
		}
	}
	go func() {
		err := s.RPC.Start()
		if err != nil {
			s.log.Fatal("unable to start rpc server")
		}
	}()
}

func (s *Server) Stop() error {
	s.Chain.Stop()
	s.RPC.Stop()
	if s.Proposer != nil {
		s.Proposer.Stop()
	}
	return nil
}

func NewServer(ctx context.Context, configParams *config.Config, logger *logger.Logger, currParams params.ChainParams, db *bdb.BlockDB, ip primitives.InitializationParameters) (*Server, error) {
	logger.Tracef("loading network parameters for '%v'", currParams.Name)
	ch, err := chain.NewBlockchain(loadChainConfig(configParams, logger), currParams, db, ip)
	if err != nil {
		return nil, err
	}
	hostnode, err := peers.NewHostNode(ctx, loadPeersManConfig(configParams, logger), ch)
	if err != nil {
		return nil, err
	}
	coinsMempool, err := mempool.NewCoinsMempool(ctx, logger, ch, hostnode, &currParams)
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
	voteMempool.Notify(actionsMempool)
	k, err := keystore.NewKeystore(configParams.DataFolder, logger)
	if err != nil {
		return nil, err
	}
	w, err := wallet.NewWallet(ctx, logger, configParams.DataFolder, &currParams, ch, hostnode, coinsMempool, actionsMempool)
	if err != nil {
		return nil, err
	}
	rpc, err := chainrpc.NewRPCServer(loadRPCConfig(configParams, logger), ch, k, hostnode, w, &currParams)
	if err != nil {
		return nil, err
	}
	var prop *proposer.Proposer
	if configParams.MiningEnabled {
		prop, err = proposer.NewProposer(loadProposerConfig(configParams, logger), currParams, ch, k, hostnode, voteMempool, coinsMempool, actionsMempool)
		if err != nil {
			return nil, err
		}
	}
	s := &Server{
		config: configParams,
		log:    logger,

		Chain:    ch,
		HostNode: hostnode,
		Keystore: k,
		Proposer: prop,
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

func loadProposerConfig(config *config.Config, logger *logger.Logger) proposer.Config {
	cfg := proposer.Config{
		Log: logger,
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
		Log:          logger,
		RPCWallet:    config.RPCWallet,
		RPCProxy:     config.RPCProxy,
		RPCProxyPort: config.RPCProxyPort,
		RPCPort:      config.RPCPort,
		Network:      "tcp",
	}
}
