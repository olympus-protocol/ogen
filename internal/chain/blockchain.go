package chain

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"sync"
	"time"

	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// Blockchain is an interface for blockchain
type Blockchain interface {
	Start() (err error)
	Stop()
	State() StateService
	GenesisTime() time.Time
	GetBlock(h chainhash.Hash) (block *primitives.Block, err error)
	GetRawBlock(h chainhash.Hash) (block []byte, err error)
	Notify(n BlockchainNotifee)
	Unnotify(n BlockchainNotifee)
	UpdateChainHead(possible chainhash.Hash, isCheck bool) error
	ProcessBlock(block *primitives.Block) error
}

var _ Blockchain = &blockchain{}

type blockchain struct {
	log         logger.Logger
	genesisTime time.Time
	netParams   *params.ChainParams

	// DB
	db blockdb.Database

	// StateService
	state StateService

	notifees    map[BlockchainNotifee]struct{}
	notifeeLock sync.Mutex
}

func (ch *blockchain) Start() (err error) {
	ch.log.Info("Starting Blockchain instance")
	return nil
}

func (ch *blockchain) Stop() {
	ch.log.Info("Stopping Blockchain instance")
}

func (ch *blockchain) State() StateService {
	return ch.state
}

func (ch *blockchain) GenesisTime() time.Time {
	return ch.genesisTime
}

// GetBlock gets a block from the database.
func (ch *blockchain) GetBlock(h chainhash.Hash) (block *primitives.Block, err error) {
	return ch.db.GetBlock(h)
}

// GetRawBlock gets the block bytes from the database.
func (ch *blockchain) GetRawBlock(h chainhash.Hash) (block []byte, err error) {
	return ch.db.GetRawBlock(h)
}

// NewBlockchain constructs a new blockchain.
func NewBlockchain(db blockdb.Database) (Blockchain, error) {

	log := config.GlobalParams.Logger
	ip := config.GlobalParams.InitParams
	netParams := config.GlobalParams.NetParams

	s, err := NewStateService(db)
	if err != nil {
		return nil, err
	}
	var genesisTime time.Time

	genesisTime, err = db.GetGenesisTime()
	if err != nil {
		log.Infof("using genesis time %d from params", ip.GenesisTime.Unix())
		genesisTime = ip.GenesisTime
		if err := db.SetGenesisTime(ip.GenesisTime); err != nil {
			return nil, err
		}
	} else {
		log.Infof("using genesis time %d from db", genesisTime.Unix())
	}
	ch := &blockchain{
		log:         log,
		netParams:   netParams,
		db:          db,
		state:       s,
		notifees:    make(map[BlockchainNotifee]struct{}),
		genesisTime: genesisTime,
	}
	return ch, ch.UpdateChainHead(s.Tip().Hash, false)
}
