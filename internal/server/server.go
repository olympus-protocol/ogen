package server

import (
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/dashboard"
	"github.com/olympus-protocol/ogen/internal/host"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/proposer"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"sync"
)

type Server interface {
	Host() host.Host
	Proposer() proposer.Proposer
	Chain() chain.Blockchain
	Start()
	Stop() error
}

// Server is the main struct that contains ogen services
type server struct {
	log logger.Logger

	lock sync.Mutex

	ch        chain.Blockchain
	h         host.Host
	prop      proposer.Proposer
	dashboard *dashboard.Dashboard
	pool      mempool.Pool

	rpcAPIs       []rpc.API
	http          *httpServer
	ws            *httpServer
	inprocHandler *rpc.Server
}

var _ Server = &server{}

func (s *server) Host() host.Host {
	return s.h
}

func (s *server) Proposer() proposer.Proposer {
	return s.prop
}

func (s *server) Chain() chain.Blockchain {
	return s.ch
}

// Start starts running the multiple ogen services.
func (s *server) Start() {

	s.pool.Start()

	err := s.ch.Start()
	if err != nil {
		s.log.Fatal("unable to start chain instance")
	}

	err = s.prop.Start()
	if err != nil {
		s.log.Fatal("unable to start proposer")
	}

	err = s.openEndpoints()
	if err != nil {
		s.log.Fatal("unable to start rpc")
	}

	if config.GlobalFlags.Dashboard {
		go func() {
			err = s.dashboard.Start()
			if err != nil {
				s.log.Fatal(err)
			}
		}()
	}
}

// openEndpoints starts all network and RPC endpoints.
func (s *server) openEndpoints() error {
	// start networking endpoints
	s.log.Info("Starting RPC Service")
	// start RPC endpoints
	err := s.startRPC()
	if err != nil {
		s.stopRPC()
	}
	return err
}

// configureRPC is a helper method to configure all the various RPC endpoints during node
// startup. It's not meant to be called at any time afterwards as it makes certain
// assumptions about the state of the node.
func (s *server) startRPC() error {
	if err := s.startInProc(); err != nil {
		return err
	}

	// Configure HTTP.
	if config.GlobalFlags.HTTPHost != "" {
		cfg := httpConfig{
			CorsAllowedOrigins: config.GlobalFlags.HTTPCors,
			Vhosts:             config.GlobalFlags.HTTPVirtualHosts,
			Modules:            config.GlobalFlags.HTTPModules,
			prefix:             config.GlobalFlags.HTTPPathPrefix,
		}
		if err := s.http.setListenAddr(config.GlobalFlags.HTTPHost, config.GlobalFlags.HTTPPort); err != nil {
			return err
		}
		if err := s.http.enableRPC(s.rpcAPIs, cfg); err != nil {
			return err
		}
	}

	// Configure WebSocket.
	if config.GlobalFlags.WSHost != "" {
		server := s.wsServerForPort(config.GlobalFlags.WSPort)
		cfg := wsConfig{
			Modules: config.GlobalFlags.WSModules,
			Origins: config.GlobalFlags.WSOrigins,
			prefix:  config.GlobalFlags.WSPathPrefix,
		}
		if err := server.setListenAddr(config.GlobalFlags.WSHost, config.GlobalFlags.WSPort); err != nil {
			return err
		}
		if err := server.enableWS(s.rpcAPIs, cfg); err != nil {
			return err
		}
	}

	if err := s.http.start(); err != nil {
		return err
	}
	return s.ws.start()
}

func (s *server) wsServerForPort(port int) *httpServer {
	if config.GlobalFlags.HTTPHost == "" || s.http.port == port {
		return s.http
	}
	return s.ws
}

// Stop closes the ogen services.
func (s *server) Stop() error {
	s.stopRPC()
	s.ch.Stop()
	s.pool.Close()
	s.h.Stop()
	return nil
}

func (s *server) stopRPC() {
	s.http.stop()
	s.ws.stop()
	s.stopInProc()
}

// startInProc registers all RPC APIs on the inproc server.
func (s *server) startInProc() error {
	for _, api := range s.rpcAPIs {
		if err := s.inprocHandler.RegisterName(api.Namespace, api.Service); err != nil {
			return err
		}
	}
	return nil
}

// stopInProc terminates the in-process RPC endpoint.
func (s *server) stopInProc() {
	s.inprocHandler.Stop()
}

// NewServer creates a server instance and initializes the ogen services.
func NewServer(db blockdb.Database) (Server, error) {

	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	log.Tracef("Loading network parameters for %v", netParams.Name)

	log.Tracef("Initializing bls module with params for %v", netParams.Name)

	ch, err := chain.NewBlockchain(db)
	if err != nil {
		return nil, err
	}

	h, err := host.NewHostNode(ch)
	if err != nil {
		return nil, err
	}

	pool := mempool.NewPool(ch, h)

	ks := keystore.NewKeystore()

	prop, err := proposer.NewProposer(ch, h, pool, ks)
	if err != nil {
		return nil, err
	}

	s := &server{
		log: log,

		ch:   ch,
		h:    h,
		prop: prop,
		pool: pool,

		inprocHandler: rpc.NewServer(),
	}

	// Register built-in APIs.
	s.rpcAPIs = append(s.rpcAPIs, s.apis()...)

	if config.GlobalFlags.Dashboard {
		s.dashboard, err = dashboard.NewDashboard(h, ch, prop)
		if err != nil {
			return nil, err
		}
	}

	// Configure RPC servers.
	s.http = newHTTPServer(s.log, rpc.DefaultHTTPTimeouts)
	s.ws = newHTTPServer(s.log, rpc.DefaultHTTPTimeouts)

	return s, nil
}
