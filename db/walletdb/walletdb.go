package walletdb

import (
	"bytes"
	"errors"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"go.etcd.io/bbolt"
)

var (
	ErrorNoMetaBucket        = errors.New("no metadata information found")
	ErrorNoCredentialsBucket = errors.New("no credentials information found")
	ErrorNoUtxosBucket       = errors.New("no utxos information found")
	ErrorNoTxsBucket         = errors.New("no transactions information found")
)

type WalletDB struct {
	db *bbolt.DB
}

func (wdb *WalletDB) Close() error {
	return wdb.db.Close()
}

func (wdb *WalletDB) GetMetadata() (*WalletMetaData, error) {
	var data []byte
	err := wdb.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(walletMetaBucketKey)
		if bkt != nil {
			data = bkt.Get([]byte("data"))
		} else {
			return ErrorNoMetaBucket
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(data)
	var metaData WalletMetaData
	err = metaData.Deserialize(buf)
	if err != nil {
		return nil, err
	}
	return &metaData, nil
}

func (wdb *WalletDB) StoreMetadata(data WalletMetaData) error {
	err := wdb.db.Update(func(tx *bbolt.Tx) error {
		buf := bytes.NewBuffer([]byte{})
		err := data.Serialize(buf)
		if err != nil {
			return err
		}
		bkt, err := tx.CreateBucketIfNotExists(walletMetaBucketKey)
		if err != nil {
			return err
		}
		err = bkt.Put([]byte("data"), buf.Bytes())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (wdb *WalletDB) GetCredentials() (*WalletCredentials, error) {
	var data []byte
	err := wdb.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(walletCredentialsBucketKey)
		if bkt != nil {
			data = bkt.Get([]byte("data"))
		} else {
			return ErrorNoCredentialsBucket
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(data)
	var metaData WalletCredentials
	err = metaData.Deserialize(buf)
	if err != nil {
		return nil, err
	}
	return &metaData, nil
}

func (wdb *WalletDB) StoreCredentials(data *WalletCredentials) error {
	err := wdb.db.Update(func(tx *bbolt.Tx) error {
		buf := bytes.NewBuffer([]byte{})
		err := data.Serialize(buf)
		if err != nil {
			return err
		}
		bkt, err := tx.CreateBucketIfNotExists(walletCredentialsBucketKey)
		if err != nil {
			return err
		}
		err = bkt.Put([]byte("data"), buf.Bytes())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (wdb *WalletDB) GetUtxos() ([]WalletUtxo, error) {
	var data [][]byte
	err := wdb.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(walletUtxosBucketKey)
		if bkt != nil {
			err := bkt.ForEach(func(k, v []byte) error {
				data = append(data, v)
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			return ErrorNoUtxosBucket
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var utxos []WalletUtxo
	for _, rawWalletUtxo := range data {
		var utxo WalletUtxo
		buf := bytes.NewBuffer(rawWalletUtxo)
		err = utxo.Deserialize(buf)
		if err != nil {
			return nil, err
		}
		utxos = append(utxos, utxo)
	}
	return utxos, nil
}

func (wdb *WalletDB) GetUtxo(utxoHash chainhash.Hash) (*WalletUtxo, error) {
	var data []byte
	err := wdb.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(walletUtxosBucketKey)
		data = bkt.Get(utxoHash.CloneBytes())
		return nil
	})
	if err != nil {
		return nil, err
	}
	var utxo WalletUtxo
	buf := bytes.NewBuffer(data)
	err = utxo.Deserialize(buf)
	if err != nil {
		return nil, err
	}
	return &utxo, nil
}

func (wdb *WalletDB) StoreUtxo(utxo WalletUtxo) error {
	err := wdb.db.Update(func(tx *bbolt.Tx) error {
		buf := bytes.NewBuffer([]byte{})
		err := utxo.Serialize(buf)
		if err != nil {
			return err
		}
		bkt, err := tx.CreateBucketIfNotExists(walletUtxosBucketKey)
		if err != nil {
			return err
		}
		utxoHash, err := utxo.OutPoint.Hash()
		if err != nil {
			return err
		}
		err = bkt.Put(utxoHash.CloneBytes(), buf.Bytes())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (wdb *WalletDB) InitUtxosBucket() error {
	err := wdb.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucket(walletUtxosBucketKey)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (wdb *WalletDB) GetTxs() ([]WalletTx, error) {
	var data [][]byte
	err := wdb.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(walletTxsBucketKey)
		if bkt != nil {
			err := bkt.ForEach(func(k, v []byte) error {
				data = append(data, v)
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			return ErrorNoTxsBucket
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var txs []WalletTx
	for _, rawWalletTx := range data {
		var tx WalletTx
		buf := bytes.NewBuffer(rawWalletTx)
		err = tx.Deserialize(buf)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (wdb *WalletDB) GetTx(TxID chainhash.Hash) (*WalletTx, error) {
	var data []byte
	err := wdb.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(walletTxsBucketKey)
		data = bkt.Get(TxID.CloneBytes())
		return nil
	})
	if err != nil {
		return nil, err
	}
	var tx WalletTx
	buf := bytes.NewBuffer(data)
	err = tx.Deserialize(buf)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (wdb *WalletDB) StoreTx(wtx WalletTx) error {
	err := wdb.db.Update(func(tx *bbolt.Tx) error {
		buf := bytes.NewBuffer([]byte{})
		err := wtx.Serialize(buf)
		if err != nil {
			return err
		}
		bkt, err := tx.CreateBucketIfNotExists(walletTxsBucketKey)
		if err != nil {
			return err
		}
		err = bkt.Put(wtx.TxID.CloneBytes(), buf.Bytes())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (wdb *WalletDB) InitTxBucket() error {
	err := wdb.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucket(walletTxsBucketKey)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func NewWalletDB(path string) *WalletDB {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		panic("unable to open wallet")
	}
	return &WalletDB{
		db: db,
	}
}
