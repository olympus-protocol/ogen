package chain

import (
	"sync"

	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
)

// StateService keeps track of the blockchain and its state. This is where pruning should eventually be implemented to
// get rid of old states.
type StateService struct {
	log      *logger.Logger
	lock     sync.RWMutex
	params   params.ChainParams

	View     *ChainView

	sync bool
}

func (s *StateService) IsSync() bool {
	return s.sync
}

func (s *StateService) SetSyncStatus(sync bool) {
	s.sync = sync
	return
}

func (s *StateService) initChainState(db blockdb.DB, params params.ChainParams) error {
	// Get the state snap from db dbindex and deserialize
	s.log.Info("loading chain state...")

	// load chain state
	loc, err := db.AddRawBlock(&params.GenesisBlock)
	if err != nil {
		return err
	}
	view, err := NewChainView(params.GenesisBlock.Header, *loc)
	if err != nil {
		return err
	}
	s.View = view

	// TODO: load block index
	return nil
}

func NewStateService(log *logger.Logger, params params.ChainParams, db blockdb.DB) (*StateService, error) {
	ss := &StateService{
		params:   params,
		log:      log,
		sync:     false,
	}
	err := ss.initChainState(db, params)
	if err != nil {
		return nil, err
	}
	return ss, nil
}
