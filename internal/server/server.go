package server

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/actionmanager"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainrpc"
	"github.com/olympus-protocol/ogen/internal/dashboard"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/internal/proposer"
	"github.com/olympus-protocol/ogen/internal/wallet"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/logger"
)

type Server interface {
	HostNode() hostnode.HostNode
	Proposer() proposer.Proposer
	Chain() chain.Blockchain
	Start()
	Stop() error
}

// Server is the main struct that contains ogen services
type server struct {
	log logger.Logger

	ch        chain.Blockchain
	hn        hostnode.HostNode
	rpc       chainrpc.RPCServer
	prop      proposer.Proposer
	dashboard *dashboard.Dashboard
}

var _ Server = &server{}

func (s *server) HostNode() hostnode.HostNode {
	return s.hn
}

func (s *server) Proposer() proposer.Proposer {
	return s.prop
}

func (s *server) Chain() chain.Blockchain {
	return s.ch
}

// Start starts running the multiple ogen services.
func (s *server) Start() {
	go func() {
		err := s.rpc.Start()
		if err != nil {
			s.log.Fatal("unable to start rpc server")
		}
	}()
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
	return nil
}

// NewServer creates a server instance and initializes the ogen services.
func NewServer(db blockdb.Database) (Server, error) {

	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	log.Tracef("Loading network parameters for %v", netParams.Name)

	log.Tracef("Initializing bls module with params for %v", netParams.Name)

	bls.Initialize(netParams)

	ch, err := chain.NewBlockchain(db)
	if err != nil {
		return nil, err
	}

	hn, err := hostnode.NewHostNode(ch)
	if err != nil {
		return nil, err
	}

	lam, err := actionmanager.NewLastActionManager(hn, ch)
	if err != nil {
		return nil, err
	}

	cpool, err := mempool.NewCoinsMempool(ch, hn)
	if err != nil {
		return nil, err
	}

	vpool, err := mempool.NewVoteMempool(ch, hn, lam)
	if err != nil {
		return nil, err
	}

	apool, err := mempool.NewActionMempool(ch, hn)
	if err != nil {
		return nil, err
	}

	vpool.Notify(apool)

	w, err := wallet.NewWallet(ch, hn, cpool, apool)
	if err != nil {
		return nil, err
	}

	ks := keystore.NewKeystore()

	prop, err := proposer.NewProposer(ch, hn, vpool, cpool, apool, lam, ks)
	if err != nil {
		return nil, err
	}

	rpc, err := chainrpc.NewRPCServer(ch, hn, w, ks, cpool, apool)
	if err != nil {
		return nil, err
	}

	s := &server{
		log: log,

		ch:   ch,
		hn:   hn,
		rpc:  rpc,
		prop: prop,
	}

	if config.GlobalFlags.Dashboard {
		s.dashboard, err = dashboard.NewDashboard(hn, ch, prop)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}
