package chainrpc

import (
	"context"
	"crypto/tls"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"net"
	"net/http"
	"path"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/wallet"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Config config for the RPCServer
type Config struct {
	datapath     string
	network      string
	rpcwallet    bool
	rpcproxy     bool
	rpcproxyport string
	rpcproxyaddr string
	rpcport      string
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
	config           *Config
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
	if s.config.rpcwallet {
		proto.RegisterWalletServer(s.rpc, s.walletServer)
	}
}

func (s *rpcServer) registerServicesProxy(ctx context.Context) {
	certPool, err := LoadCerts()
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
	if s.config.rpcwallet {
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

	if s.config.rpcproxy {

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		s.registerServicesProxy(ctx)

		go func() {
			var addr string
			if s.config.rpcproxyaddr != "" {
				addr = s.config.rpcproxyaddr
			} else {
				addr = "localhost"
			}
			c := cors.New(cors.Options{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{http.MethodGet, http.MethodPost},
			})
			handler := c.Handler(s.http)
			err := http.ListenAndServeTLS(addr+":"+s.config.rpcproxyport, path.Join(config.GlobalFlags.DataPath, "cert", "cert.pem"), path.Join(config.GlobalFlags.DataPath, "cert", "cert_key.pem"), handler)
			if err != nil {
				s.log.Fatal(err)
			}

		}()
	}
	lis, err := net.Listen("tcp", "127.0.0.1:"+s.config.rpcport)
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
func NewRPCServer(chain chain.Blockchain, hostnode hostnode.HostNode, wallet wallet.Wallet, ks keystore.Keystore, cm mempool.CoinsMempool, am mempool.ActionMempool) (RPCServer, error) {
	datapath := config.GlobalFlags.DataPath
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	_, err := LoadCerts()
	if err != nil {
		return nil, err
	}
	creds, err := credentials.NewServerTLSFromFile(path.Join(datapath, "cert", "cert.pem"), path.Join(datapath, "cert", "cert_key.pem"))
	if err != nil {
		return nil, err
	}
	return &rpcServer{
		rpc:  grpc.NewServer(grpc.Creds(creds)),
		http: runtime.NewServeMux(),
		config: &Config{
			datapath:     config.GlobalFlags.DataPath,
			network:      config.GlobalFlags.NetworkName,
			rpcwallet:    config.GlobalFlags.RPCWallet,
			rpcproxy:     config.GlobalFlags.RPCProxy,
			rpcproxyport: config.GlobalFlags.RPCProxyPort,
			rpcproxyaddr: config.GlobalFlags.RPCProxyAddr,
			rpcport:      config.GlobalFlags.RPCPort,
		},
		log: log,
		chainServer: &chainServer{
			chain: chain,
		},
		validatorsServer: &validatorsServer{
			netParams: netParams,
			chain:     chain,
		},
		networkServer: &networkServer{
			host: hostnode,
		},
		utilsServer: &utilsServer{
			keystore:       ks,
			host:           hostnode,
			coinsMempool:   cm,
			chain:          chain,
			actionsMempool: am,
		},
		walletServer: &walletServer{
			wallet:    wallet,
			chain:     chain,
			netParams: netParams,
		},
	}, nil
}
