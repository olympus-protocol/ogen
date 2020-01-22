package walletdb

import (
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/utils/chainhash"
	"github.com/grupokindynos/ogen/utils/serializer"
	"io"
	"time"
)

var walletTxsBucketKey = []byte("wallet-txs")

type WalletTx struct {
	Type      p2p.TxType
	Action    p2p.TxAction
	TxID      chainhash.Hash
	Value     int64
	Spent     bool
	Timestamp time.Time
}

func (tx *WalletTx) Serialize(w io.Writer) error {
	timestamp := uint32(tx.Timestamp.Unix())
	err := serializer.WriteElements(w, tx.Type, tx.Action, tx.TxID, tx.Value, timestamp)
	if err != nil {
		return err
	}
	return nil
}

func (tx *WalletTx) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &tx.Type, &tx.Action, &tx.TxID, &tx.Value, (*serializer.Uint32Time)(&tx.Timestamp))
	if err != nil {
		return err
	}
	return nil
}
