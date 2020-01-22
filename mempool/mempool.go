package mempool

import (
	"github.com/grupokindynos/ogen/logger"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/params"
	"github.com/grupokindynos/ogen/txs/txverifier"
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

func InitMempool(config Config, txverifier *txverifier.TxVerifier, params params.ChainParams) *Mempool {
	return &Mempool{
		config:     config,
		log:        config.Log,
		params:     params,
		Tx:         []*p2p.MsgTx{},
		txverifier: txverifier,
	}
}
