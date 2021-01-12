package server

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainrpc"
	"github.com/olympus-protocol/ogen/internal/dashboard"
	"github.com/olympus-protocol/ogen/internal/host"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/proposer"
	"github.com/olympus-protocol/ogen/internal/wallet"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/logger"
)

type Server interface {
	Host() host.Host
	Proposer() proposer.Proposer
	Chain() chain.Blockchain
	Start()
	Stop() error
	Wallet() wallet.Wallet
}

// Server is the main struct that contains ogen services
type server struct {
	log logger.Logger

	ch        chain.Blockchain
	h         host.Host
	rpc       chainrpc.RPCServer
	prop      proposer.Proposer
	dashboard *dashboard.Dashboard
	wallet    wallet.Wallet
	pool      mempool.Pool
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

func (s *server) Wallet() wallet.Wallet {
	return s.wallet
}

// Start starts running the multiple ogen services.
func (s *server) Start() {
	go func() {
		err := s.rpc.Start()
		if err != nil {
			s.log.Fatal("unable to start rpc server")
		}
	}()

	s.pool.Start()

	err := s.ch.Start()
	if err != nil {
		s.log.Fatal("unable to start chain instance")
	}

	err = s.prop.Start()
	if err != nil {
		s.log.Fatal("unable to start proposer")
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

// Stop closes the ogen services.
func (s *server) Stop() error {
	s.ch.Stop()
	s.rpc.Stop()
	s.pool.Close()
	s.h.Stop()
	return nil
}

// NewServer creates a server instance and initializes the ogen services.
func NewServer(db blockdb.Database) (Server, error) {

	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	log.Tracef("Loading network parameters for %v", netParams.Name)

	log.Tracef("Initializing bls module with params for %v", netParams.Name)

	bls.Initialize(netParams, "blst")

	ch, err := chain.NewBlockchain(db)
	if err != nil {
		return nil, err
	}

	h, err := host.NewHostNode(ch)
	if err != nil {
		return nil, err
	}

	//lam, err := actionmanager.NewLastActionManager(h, ch)
	//if err != nil {
	//	return nil, err
	//}

	pool := mempool.NewPool(ch, h)

	w, err := wallet.NewWallet(ch, h, pool)
	if err != nil {
		return nil, err
	}

	ks := keystore.NewKeystore()

	prop, err := proposer.NewProposer(ch, h, pool, ks)
	if err != nil {
		return nil, err
	}

	rpc, err := chainrpc.NewRPCServer(ch, h, w, ks, pool)
	if err != nil {
		return nil, err
	}

	s := &server{
		log: log,

		ch:     ch,
		h:      h,
		rpc:    rpc,
		prop:   prop,
		wallet: w,
		pool:   pool,
	}

	if config.GlobalFlags.Dashboard {
		s.dashboard, err = dashboard.NewDashboard(h, ch, prop)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}
