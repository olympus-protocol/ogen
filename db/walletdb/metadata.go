package walletdb

import (
	"github.com/grupokindynos/ogen/utils/chainhash"
	"github.com/grupokindynos/ogen/utils/serializer"
	"io"
)

var walletMetaBucketKey = []byte("wallet-metadata")

type WalletMetaData struct {
	Version         int64
	Txs             int64
	Utxos           int64
	Accounts        int64
	LastBlockHash   chainhash.Hash
	LastBlockHeight int64
}

func (meta *WalletMetaData) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, meta.Version, meta.Txs, meta.Utxos, meta.Accounts, meta.LastBlockHash, meta.LastBlockHeight)
	if err != nil {
		return err
	}
	return nil
}

func (meta *WalletMetaData) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &meta.Version, &meta.Txs, &meta.Utxos, &meta.Accounts, &meta.LastBlockHash, &meta.LastBlockHeight)
	if err != nil {
		return err
	}
	return nil
}
