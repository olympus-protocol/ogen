package mempool

import (
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/txs/txverifier"
)

type Config struct {
	Log *logger.Logger
}

type Mempool struct {
	config     Config
	log        *logger.Logger
	params     params.ChainParams
	txverifier *txverifier.TxVerifier

	Tx []*p2p.MsgTx
}

func (m *Mempool) GetTxs() []*p2p.MsgTx {
	return m.Tx
}

func (m *Mempool) AddTx(tx *p2p.MsgTx) error {
	err := m.txverifier.VerifyTx(tx)
	if err != nil {
		return err
	}
	m.Tx = append(m.Tx, tx)
	return nil
}

func InitMempool(config Config, params params.ChainParams) *Mempool {
	return &Mempool{
		config: config,
		log:    config.Log,
		params: params,
		Tx:     []*p2p.MsgTx{},
	}
}
