package chainrpc

import (
	"net"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chainrpc/proto"
	"github.com/olympus-protocol/ogen/utils/logger"
	"github.com/olympus-protocol/ogen/wallet"
	"google.golang.org/grpc"
)

type Config struct {
	Network string
	Address string
	Log     *logger.Logger
}

type RPCServer struct {
	log    *logger.Logger
	config Config
	rpc    *grpc.Server

	chainServer      *chainServer
	validatorsServer *validatorsServer
}

func (s *RPCServer) registerServices() {
	proto.RegisterChainServer(s.rpc, s.chainServer)
	proto.RegisterValidatorsServer(s.rpc, s.validatorsServer)
}

func (s *RPCServer) Stop() {
	s.log.Info("stoping gRPC Server")
	s.rpc.GracefulStop()
}

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
func NewRPCServer(config Config, chain *chain.Blockchain, wallet *wallet.Wallet) *RPCServer {
	return &RPCServer{
		rpc:    grpc.NewServer(),
		config: config,
		log:    config.Log,
		chainServer: &chainServer{
			chain: chain,
		},
		validatorsServer: &validatorsServer{
			wallet: wallet,
			chain:  chain,
		},
	}
}
