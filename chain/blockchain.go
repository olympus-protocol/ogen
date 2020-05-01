package chain

import (
	"sync"
	"time"

	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/mempool"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
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
	state   *StateService
	mempool *mempool.CoinsMempool

	notifees    map[BlockchainNotifee]struct{}
	notifeeLock sync.RWMutex
}

func (ch *Blockchain) SubmitCoinTransaction(tx *primitives.CoinPayload) error {
	state := ch.state.TipState()
	return ch.mempool.Add(*tx, &state.UtxoState)
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

// GetBlock gets a block from the database.
func (ch *Blockchain) GetBlock(h chainhash.Hash) (block *primitives.Block, err error) {
	return block, ch.db.View(func(txn blockdb.DBViewTransaction) error {
		block, err = txn.GetRawBlock(h)
		return err
	})
}

// NewBlockchain constructs a new blockchain.
func NewBlockchain(config Config, params params.ChainParams, db blockdb.DB, ip primitives.InitializationParameters, m *mempool.CoinsMempool) (*Blockchain, error) {
	state, err := NewStateService(config.Log, ip, params, db)
	if err != nil {
		return nil, err
	}
	var genesisTime time.Time

	err = db.Update(func(tx blockdb.DBUpdateTransaction) error {
		genesisTime, err = tx.GetGenesisTime()
		if err != nil {
			config.Log.Infof("using genesis time %d from params", ip.GenesisTime.Unix())
			genesisTime = ip.GenesisTime
			if err := tx.SetGenesisTime(ip.GenesisTime); err != nil {
				return err
			}
		} else {
			config.Log.Infof("using genesis time %d from db", genesisTime.Unix())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ch := &Blockchain{
		log:         config.Log,
		config:      config,
		params:      params,
		db:          db,
		state:       state,
		notifees:    make(map[BlockchainNotifee]struct{}),
		genesisTime: genesisTime,
		mempool:     m,
	}
	return ch, db.Update(func(txn blockdb.DBUpdateTransaction) error {
		return ch.UpdateChainHead(txn)
	})
}
