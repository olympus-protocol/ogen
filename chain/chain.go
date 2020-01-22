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

type Chain struct {
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

func (ch *Chain) Start() (err error) {
	ch.log.Info("Starting Chain instance")
	err = ch.state.InitChainState(ch.db, ch.params)
	if err != nil {
		return err
	}
	ch.log.Infof(ch.state.snapshot.String())
	return nil
}

func (ch *Chain) Stop() {
	ch.log.Info("Stoping Chain instance")
}

func (ch *Chain) StateSnapshot() *StateSnap {
	return ch.state.Snapshot()
}

func (ch *Chain) State() *State {
	return ch.state
}

func (ch *Chain) UpdateState(block *primitives.Block, workers int64, users int64, govObjects int64, store bool) error {
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

func NewChain(config Config, params params.ChainParams, indexers *index.Indexers, txverifier *txverifier.TxVerifier, db *blockdb.BlockDB) (*Chain, error) {
	ch := &Chain{
		log:        config.Log,
		config:     config,
		params:     params,
		db:         db,
		state:      NewChainState(indexers, config.Log, params),
		txverifier: txverifier,
	}
	return ch, nil
}
