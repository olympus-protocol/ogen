package chainrpc

import (
	"context"
	"crypto/tls"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"net"
	"net/http"
	"path"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/proposer"
	"github.com/olympus-protocol/ogen/pkg/params"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/wallet"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Config config for the RPCServer
type Config struct {
	DataDir      string
	Network      string
	RPCWallet    bool
	RPCProxy     bool
	RPCProxyPort string
	RPCProxyAddr string
	RPCPort      string
	Log          logger.Logger
}

//RPCServer is an interface for rpcServer
type RPCServer interface {
	Stop()
	Start() error
}

var _ RPCServer = &rpcServer{}

// rpcServer struct model for the gRPC server
type rpcServer struct {
	log              logger.Logger
	config           Config
	http             *runtime.ServeMux
	rpc              *grpc.Server
	chainServer      *chainServer
	validatorsServer *validatorsServer
	utilsServer      *utilsServer
	networkServer    *networkServer
	walletServer     *walletServer
}

func (s *rpcServer) registerServices() {
	proto.RegisterChainServer(s.rpc, s.chainServer)
	proto.RegisterValidatorsServer(s.rpc, s.validatorsServer)
	proto.RegisterUtilsServer(s.rpc, s.utilsServer)
	proto.RegisterNetworkServer(s.rpc, s.networkServer)
	if s.config.RPCWallet {
		proto.RegisterWalletServer(s.rpc, s.walletServer)
	}
}

func (s *rpcServer) registerServicesProxy(ctx context.Context) {
	certPool, err := LoadCerts(s.config.DataDir)
	if err != nil {
		s.log.Fatal(err)
	}
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            certPool,
	})
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	err = proto.RegisterChainHandlerFromEndpoint(ctx, s.http, "127.0.0.1:24127", opts)
	if err != nil {
		s.log.Fatal(err)
	}
	err = proto.RegisterValidatorsHandlerFromEndpoint(ctx, s.http, "127.0.0.1:24127", opts)
	if err != nil {
		s.log.Fatal(err)
	}
	err = proto.RegisterUtilsHandlerFromEndpoint(ctx, s.http, "127.0.0.1:24127", opts)
	if err != nil {
		s.log.Fatal(err)
	}
	err = proto.RegisterNetworkHandlerFromEndpoint(ctx, s.http, "127.0.0.1:24127", opts)
	if err != nil {
		s.log.Fatal(err)
	}
	if s.config.RPCWallet {
		err = proto.RegisterWalletHandlerFromEndpoint(ctx, s.http, "127.0.0.1:24127", opts)
		if err != nil {
			s.log.Fatal(err)
		}
	}
}

// Stop stops gRPC listener
func (s *rpcServer) Stop() {
	s.log.Info("Stopping gRPC Server")
	s.rpc.GracefulStop()
}

// Start starts gRPC listener
func (s *rpcServer) Start() error {
	s.registerServices()
	s.log.Info("Starting gRPC Server")
	if s.config.RPCProxy {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		s.registerServicesProxy(ctx)
		go func() {
			var addr string
			if s.config.RPCProxyAddr != "" {
				addr = s.config.RPCProxyAddr
			} else {
				addr = "localhost"
			}
			c := cors.New(cors.Options{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{http.MethodGet, http.MethodPost},
			})
			handler := c.Handler(s.http)
			err := http.ListenAndServeTLS(addr+":"+s.config.RPCProxyPort, path.Join(s.config.DataDir, "cert", "cert.pem"), path.Join(s.config.DataDir, "cert", "cert_key.pem"), handler)
			if err != nil {
				s.log.Fatal(err)
			}
		}()
	}
	lis, err := net.Listen("tcp", "127.0.0.1:"+s.config.RPCPort)
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
func NewRPCServer(config Config, chain chain.Blockchain, hostnode hostnode.HostNode, wallet wallet.Wallet, params *params.ChainParams, p proposer.Proposer) (RPCServer, error) {
	txTopic, err := hostnode.Topic(p2p.MsgTxCmd)
	if err != nil {
		return nil, err
	}

	depositTopic, err := hostnode.Topic(p2p.MsgDepositCmd)
	if err != nil {
		return nil, err
	}

	exitTopic, err := hostnode.Topic(p2p.MsgExitCmd)
	if err != nil {
		return nil, err
	}
	_, err = LoadCerts(config.DataDir)
	if err != nil {
		return nil, err
	}
	creds, err := credentials.NewServerTLSFromFile(path.Join(config.DataDir, "cert", "cert.pem"), path.Join(config.DataDir, "cert", "cert_key.pem"))
	if err != nil {
		return nil, err
	}

	return &rpcServer{
		rpc:    grpc.NewServer(grpc.Creds(creds)),
		http:   runtime.NewServeMux(),
		config: config,
		log:    config.Log,
		chainServer: &chainServer{
			chain: chain,
		},
		validatorsServer: &validatorsServer{
			params: params,
			chain:  chain,
		},
		networkServer: &networkServer{
			hostnode: hostnode,
		},
		utilsServer: &utilsServer{
			txTopic:      txTopic,
			depositTopic: depositTopic,
			exitTopic:    exitTopic,
			proposer:     p,
			hostnode:     hostnode,
		},
		walletServer: &walletServer{
			wallet: wallet,
			chain:  chain,
			params: params,
		},
	}, nil
}
