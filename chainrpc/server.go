package chainrpc

import (
	"context"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chainrpc/proto"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/utils/logger"
	"github.com/olympus-protocol/ogen/wallet"
	"google.golang.org/grpc"
)

// Config config for the RPCServer
type Config struct {
	Network          string
	Wallet           bool
	RPCProxy         bool
	RPCListenAddress string
	Log              *logger.Logger
}

// RPCServer struct model for the gRPC server
type RPCServer struct {
	log              *logger.Logger
	config           Config
	http             *runtime.ServeMux
	rpc              *grpc.Server
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
	if s.config.Wallet {
		proto.RegisterWalletServer(s.rpc, s.walletServer)
	}
}

func (s *RPCServer) registerServicesProxy(ctx context.Context) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	proto.RegisterChainHandlerFromEndpoint(ctx, s.http, "127.0.0.1:24127", opts)
	proto.RegisterValidatorsHandlerFromEndpoint(ctx, s.http, "127.0.0.1:24127", opts)
	proto.RegisterUtilsHandlerFromEndpoint(ctx, s.http, "127.0.0.1:24127", opts)
	proto.RegisterNetworkHandlerFromEndpoint(ctx, s.http, "127.0.0.1:24127", opts)
	if s.config.Wallet {
		proto.RegisterWalletHandlerFromEndpoint(ctx, s.http, "127.0.0.1:24127", opts)
	}
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
	if s.config.RPCProxy {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		s.registerServicesProxy(ctx)
		go func() {
			err := http.ListenAndServe(":8080", s.http)
			if err != nil {
				s.log.Fatal(err)
			}
		}()
	}
	lis, err := net.Listen("tcp", s.config.RPCListenAddress)
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
func NewRPCServer(config Config, chain *chain.Blockchain, keys *keystore.Keystore, hostnode *peers.HostNode, wallet *wallet.Wallet, params *params.ChainParams) (*RPCServer, error) {
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
		http:   runtime.NewServeMux(),
		config: config,
		log:    config.Log,
		chainServer: &chainServer{
			chain: chain,
		},
		validatorsServer: &validatorsServer{
			params:   params,
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
