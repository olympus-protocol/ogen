package chain

import (
	"sync"
	"time"

	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
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
	log         *logger.Logger
	config      Config
	genesisTime time.Time
	params      params.ChainParams

	// DB
	db blockdb.DB

	// StateService
	state *StateService

	notifees    map[BlockchainNotifee]struct{}
	notifeeLock sync.RWMutex
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

func (ch *Blockchain) GenesisTime() time.Time {
	return ch.genesisTime
}

// NewBlockchain constructs a new blockchain.
func NewBlockchain(config Config, params params.ChainParams, db blockdb.DB, ip primitives.InitializationParameters) (*Blockchain, error) {
	state, err := NewStateService(config.Log, ip, params, db)
	if err != nil {
		return nil, err
	}
	var genesisTime time.Time
	genesisTime, err = db.GetGenesisTime()
	if err != nil {
		config.Log.Debugf("using genesis time %s from params", ip.GenesisTime)
		genesisTime = ip.GenesisTime
		if err := db.SetGenesisTime(ip.GenesisTime); err != nil {
			return nil, err
		}
	} else {
		config.Log.Debugf("using genesis time %s from db", genesisTime)
	}
	ch := &Blockchain{
		log:         config.Log,
		config:      config,
		params:      params,
		db:          db,
		state:       state,
		notifees:    make(map[BlockchainNotifee]struct{}),
		genesisTime: genesisTime,
	}
	return ch, nil
}
