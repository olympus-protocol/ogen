package chainrpc

import (
	"net"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/proto"
	"github.com/olympus-protocol/ogen/utils/logger"
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
}

func (s *RPCServer) registerServices() {
	proto.RegisterChainServer(s.rpc, s.chainServer)
	proto.RegisterValidatorsServer(s.rpc, s.validatorsServer)
	proto.RegisterUtilsServer(s.rpc, s.utilsServer)
	proto.RegisterNetworkServer(s.rpc, s.networkServer)

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
func NewRPCServer(config Config, chain *chain.Blockchain, keys *keystore.Keystore, host *peers.HostNode) *RPCServer {
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
			host: host,
		},
	}
}
