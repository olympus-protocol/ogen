package chain

import (
	"bytes"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/txs/txverifier"
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
	db *blockdb.BlockDB
	// State
	state      *State
	txverifier *txverifier.TxVerifier
}

func (ch *Blockchain) Start() (err error) {
	ch.log.Info("Starting Blockchain instance")
	ch.log.Infof(ch.state.snapshot.String())
	return nil
}

func (ch *Blockchain) Stop() {
	ch.log.Info("Stoping Blockchain instance")
}

func (ch *Blockchain) StateSnapshot() *StateSnap {
	return ch.state.Snapshot()
}

func (ch *Blockchain) State() *State {
	return ch.state
}

func (ch *Blockchain) UpdateState(block *primitives.Block, workers int64, users int64, govObjects int64, store bool) error {
	err := ch.state.updateStateSnap(block, workers, users, govObjects)
	if err != nil {
		return err
	}
	// TODO here we can update indexes.
	snap := ch.state.Snapshot()
	if store {
		buf := bytes.NewBuffer([]byte{})
		err = snap.Serialize(buf)
		if err != nil {
			return err
		}
		err = ch.db.SetStateSnap(buf.Bytes())
		if err != nil {
			return err
		}
	}
	ch.log.Infof(snap.String())
	return nil
}

func NewBlockchain(config Config, params params.ChainParams, indexers *index.Indexers, txverifier *txverifier.TxVerifier, db *blockdb.BlockDB) (*Blockchain, error) {
	state, err := NewChainState(indexers, config.Log, params, db)
	if err != nil {
		return nil, err
	}
	ch := &Blockchain{
		log:        config.Log,
		config:     config,
		params:     params,
		db:         db,
		state:      state,
		txverifier: txverifier,
	}
	return ch, nil
}
