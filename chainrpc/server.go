package chainrpc

import (
	"net"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/chainrpc/proto"
	"github.com/olympus-protocol/ogen/utils/logger"
	"github.com/olympus-protocol/ogen/wallet"
	"google.golang.org/grpc"
)

// Config config for the RPCServer
type Config struct {
	Network string
	Address string
	Log     *logger.Logger
}

// RPCServer struct model for the gRPC server
type RPCServer struct {
	log    *logger.Logger
	config Config
	rpc    *grpc.Server

	chainServer      *chainServer
	validatorsServer *validatorsServer
	utilsServer      *utilsServer
	networkServer    *networkServer
	walletServer     *walletServer
}

func (s *RPCServer) registerServices() {
	proto.RegisterChainServer(s.rpc, s.chainServer)
	proto.RegisterValidatorsServer(s.rpc, s.validatorsServer)
	proto.RegisterUtilsServer(s.rpc, s.utilsServer)
	proto.RegisterNetworkServer(s.rpc, s.networkServer)
	proto.RegisterWalletServer(s.rpc, s.walletServer)
}

// Stop stops gRPC listener
func (s *RPCServer) Stop() {
	s.log.Info("stoping gRPC Server")
	s.rpc.GracefulStop()
}

// Start starts gRPC listener
func (s *RPCServer) Start() error {
	s.registerServices()
	s.log.Info("Starting gRPC Server")
	lis, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return err
	}
	err = s.rpc.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}

// NewRPCServer Returns an RPC server instance
func NewRPCServer(config Config, chain *chain.Blockchain, keys *keystore.Keystore, hostnode *peers.HostNode, wallet *wallet.Wallet) (*RPCServer, error) {
	txTopic, err := hostnode.Topic("tx")
	if err != nil {
		return nil, err
	}

	depositTopic, err := hostnode.Topic("deposits")
	if err != nil {
		return nil, err
	}

	exitTopic, err := hostnode.Topic("exits")
	if err != nil {
		return nil, err
	}

	return &RPCServer{
		rpc:    grpc.NewServer(),
		config: config,
		log:    config.Log,
		chainServer: &chainServer{
			chain: chain,
		},
		validatorsServer: &validatorsServer{
			keystore: keys,
			chain:    chain,
		},
		networkServer: &networkServer{
			hostnode: hostnode,
		},
		utilsServer: &utilsServer{
			keystore:     keys,
			txTopic:      txTopic,
			depositTopic: depositTopic,
			exitTopic:    exitTopic,
		},
		walletServer: &walletServer{
			wallet: wallet,
		},
	}, nil
}
