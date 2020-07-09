package index

import (
	"path"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
	"go.etcd.io/bbolt"
)

// AccountTxs is just a helper struct for database storage of account transactions.
type AccountTxs struct {
	Amount uint64
	Txs    []chainhash.Hash
}

// Marshal encodes the data.
func (ac *AccountTxs) Marshal() ([]byte, error) {
	return ssz.Marshal(ac)
}

// Unmarshal decodes the data.
func (ac *AccountTxs) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, ac)
}

// TxIndex is a pseudo index that contains locators for account transactions.
type TxIndex struct {
	db *bbolt.DB
}

var accBucketPrefix = []byte("acc-")

// GetAccountTxs returns a list of transaction related to an account
func (i *TxIndex) GetAccountTxs(account [20]byte) (AccountTxs, error) {
	txs := AccountTxs{
		Amount: 0,
		Txs:    []chainhash.Hash{},
	}
	err := i.db.View(func(tx *bbolt.Tx) error {
		accBkt := tx.Bucket(append(accBucketPrefix, account[:]...))
		if accBkt == nil {
			return nil
		}
		err := accBkt.ForEach(func(k, _ []byte) error {
			txs.Amount++
			h, err := chainhash.NewHash(k)
			if err != nil {
				return err
			}
			txs.Txs = append(txs.Txs, *h)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return AccountTxs{}, err
	}
	return txs, nil
}

var txBucketKey = []byte("txs")

// GetTx returns a transaction locator from the index.
func (i *TxIndex) GetTx(hash chainhash.Hash) (TxLocator, error) {
	var loc TxLocator
	err := i.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(txBucketKey)
		data := bkt.Get(hash[:])
		err := loc.Unmarshal(data)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return TxLocator{}, err
	}
	return loc, nil
}

// SetTx stores a transaction locator and adds a reference to the account.
func (i *TxIndex) SetTx(locator TxLocator, account [20]byte) error {
	err := i.db.Update(func(tx *bbolt.Tx) error {
		txbkt, err := tx.CreateBucketIfNotExists(txBucketKey)
		if err != nil {
			return err
		}
		locB, err := locator.Marshal()
		if err != nil {
			return err
		}
		err = txbkt.Put(locator.Hash[:], locB)
		if err != nil {
			return err
		}
		accBkt, err := tx.CreateBucketIfNotExists(append(accBucketPrefix, account[:]...))
		if err != nil {
			return err
		}
		err = accBkt.Put(locator.Hash.CloneBytes(), []byte{})
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

// TxLocator is a simple struct to find a database referenced to a block without building a full index
type TxLocator struct {
	Hash  chainhash.Hash
	Block chainhash.Hash
	Index uint32
}

// Marshal encodes the data.
func (tl *TxLocator) Marshal() ([]byte, error) {
	return ssz.Marshal(tl)
}

// Unmarshal decodes the data.
func (tl *TxLocator) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, tl)
}

// NewTxIndex returns/creates a new tx index database
func NewTxIndex(datadir string) (*TxIndex, error) {
	db, err := bbolt.Open(path.Join(datadir, "tx.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	return &TxIndex{db: db}, nil
}
