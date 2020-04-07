package chain

import (
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
)

type BlockInfo struct {
	Height       int32
	Hash         string
	Timestamp    string
	Transactions int
	Size         uint32
}

type Config struct {
	Log *logger.Logger
}

type Blockchain struct {
	// Initial Ogen Params
	log    *logger.Logger
	config Config
	params params.ChainParams

	// DB
	db blockdb.DB

	// StateService
	state *StateService
}

func (ch *Blockchain) Start() (err error) {
	ch.log.Info("Starting Blockchain instance")
	return nil
}

func (ch *Blockchain) Stop() {
	ch.log.Info("Stoping Blockchain instance")
}

func (ch *Blockchain) State() *StateService {
	return ch.state
}

func NewBlockchain(config Config, params params.ChainParams, db blockdb.DB, ip InitializationParameters) (*Blockchain, error) {
	state, err := NewStateService(config.Log, ip, params, db)
	if err != nil {
		return nil, err
	}
	ch := &Blockchain{
		log:    config.Log,
		config: config,
		params: params,
		db:     db,
		state:  state,
	}
	return ch, nil
}
