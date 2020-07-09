package chain

import (
	"sync"
	"time"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/logger"
)

type Config struct {
	Datadir string
	Log     *logger.Logger
}

type Blockchain struct {
	// Initial Ogen Params
	log         *logger.Logger
	config      Config
	genesisTime time.Time
	params      params.ChainParams

	// DB
	db bdb.DB

	// Indexes
	txidx *index.TxIndex

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

// GetBlock gets a block from the database.
func (ch *Blockchain) GetBlock(h chainhash.Hash) (block *primitives.Block, err error) {
	return block, ch.db.View(func(txn bdb.DBViewTransaction) error {
		block, err = txn.GetBlock(h)
		return err
	})
}

// GetRawBlock gets the block bytes from the database.
func (ch *Blockchain) GetRawBlock(h chainhash.Hash) (block []byte, err error) {
	return block, ch.db.View(func(txn bdb.DBViewTransaction) error {
		block, err = txn.GetRawBlock(h)
		return err
	})
}

// GetAccountTxs gets the txid from an account.
func (ch *Blockchain) GetAccountTxs(acc [20]byte) (accTxs index.AccountTxs, err error) {
	return ch.txidx.GetAccountTxs(acc)
}

// GetTx gets the transaction from the database and block reference.
func (ch *Blockchain) GetTx(h chainhash.Hash) (tx primitives.Tx, err error) {
	loc, err := ch.txidx.GetTx(h)
	if err != nil {
		return primitives.Tx{}, err
	}
	block, err := ch.GetBlock(loc.Block)
	if err != nil {
		return primitives.Tx{}, err
	}
	return block.Txs[loc.Index], nil
}

// NewBlockchain constructs a new blockchain.
func NewBlockchain(config Config, params params.ChainParams, db bdb.DB, ip primitives.InitializationParameters) (*Blockchain, error) {
	state, err := NewStateService(config.Log, ip, params, db)
	if err != nil {
		return nil, err
	}
	var genesisTime time.Time

	err = db.Update(func(tx bdb.DBUpdateTransaction) error {
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
	txidx, err := index.NewTxIndex(config.Datadir)
	if err != nil {
		return nil, err
	}
	ch := &Blockchain{
		log:         config.Log,
		config:      config,
		params:      params,
		txidx:       txidx,
		db:          db,
		state:       state,
		notifees:    make(map[BlockchainNotifee]struct{}),
		genesisTime: genesisTime,
	}
	return ch, db.Update(func(txn bdb.DBUpdateTransaction) error {
		return ch.UpdateChainHead(txn, state.Tip().Hash)
	})
}
