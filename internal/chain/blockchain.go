package chain

import (
	"github.com/olympus-protocol/ogen/internal/txindex"
	"sync"
	"time"

	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type Config struct {
	Datadir string
	Log     logger.Logger
}

// Blockchain is an interface for blockchain
type Blockchain interface {
	Start() (err error)
	Stop()
	State() *StateService
	GenesisTime() time.Time
	GetBlock(h chainhash.Hash) (block *primitives.Block, err error)
	GetRawBlock(h chainhash.Hash) (block []byte, err error)
	GetAccountTxs(acc [20]byte) (accTxs txindex.AccountTxs, err error)
	GetTx(h chainhash.Hash) (tx *primitives.Tx, err error)
	GetLocatorHashes() [][32]byte
	Notify(n BlockchainNotifee)
	Unnotify(n BlockchainNotifee)
	UpdateChainHead(txn blockdb.DBUpdateTransaction, possible chainhash.Hash) error
	ProcessBlock(block *primitives.Block) error
}

var _ Blockchain = &blockchain{}

type blockchain struct {
	// Initial Ogen Params
	log         logger.Logger
	config      Config
	genesisTime time.Time
	params      params.ChainParams

	// DB
	db blockdb.DB

	// Indexes
	txidx *txindex.TxIndex

	// StateService
	state *StateService

	notifees    map[BlockchainNotifee]struct{}
	notifeeLock sync.RWMutex
}

func (ch *blockchain) Start() (err error) {
	ch.log.Info("Starting Blockchain instance")
	return nil
}

func (ch *blockchain) Stop() {
	ch.log.Info("Stopping Blockchain instance")
}

func (ch *blockchain) State() *StateService {
	return ch.state
}

func (ch *blockchain) GenesisTime() time.Time {
	return ch.genesisTime
}

// GetBlock gets a block from the database.
func (ch *blockchain) GetBlock(h chainhash.Hash) (block *primitives.Block, err error) {
	return block, ch.db.View(func(txn blockdb.DBViewTransaction) error {
		block, err = txn.GetBlock(h)
		return err
	})
}

// GetRawBlock gets the block bytes from the database.
func (ch *blockchain) GetRawBlock(h chainhash.Hash) (block []byte, err error) {
	return block, ch.db.View(func(txn blockdb.DBViewTransaction) error {
		block, err = txn.GetRawBlock(h)
		return err
	})
}

// GetAccountTxs gets the txid from an account.
func (ch *blockchain) GetAccountTxs(acc [20]byte) (accTxs txindex.AccountTxs, err error) {
	return ch.txidx.GetAccountTxs(acc)
}

// GetTx gets the transaction from the database and block reference.
func (ch *blockchain) GetTx(h chainhash.Hash) (tx *primitives.Tx, err error) {
	loc, err := ch.txidx.GetTx(h)
	if err != nil {
		return nil, err
	}
	block, err := ch.GetBlock(loc.Block)
	if err != nil {
		return nil, err
	}
	return block.Txs[loc.Index], nil
}

// NewBlockchain constructs a new blockchain.
func NewBlockchain(config Config, params params.ChainParams, db blockdb.DB, ip primitives.InitializationParameters) (Blockchain, error) {
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
	txidx, err := txindex.NewTxIndex(config.Datadir)
	if err != nil {
		return nil, err
	}
	ch := &blockchain{
		log:         config.Log,
		config:      config,
		params:      params,
		txidx:       txidx,
		db:          db,
		state:       state,
		notifees:    make(map[BlockchainNotifee]struct{}),
		genesisTime: genesisTime,
	}
	return ch, db.Update(func(txn blockdb.DBUpdateTransaction) error {
		return ch.UpdateChainHead(txn, state.Tip().Hash)
	})
}
