package txindex

import (
	"errors"
	"path"

	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"go.etcd.io/bbolt"
)

var (
	// ErrorTxLocatorSize returns when serialized TxLocator size exceed MaxTxLocatorSize.
	ErrorCombinedSignatureSize = errors.New("tx locator too big")
)

const (
	// MaxTxLocatorSize is the maximum amount of bytes a TxLocator can contain.
	MaxTxLocatorSize = 72
)

// AccountTxs is just a helper struct for database storage of account transactions.
type AccountTxs struct {
	Amount uint64
	Txs    []chainhash.Hash
}

// Strings returns the array of hashes on string format
func (txs *AccountTxs) Strings() []string {
	str := make([]string, txs.Amount)
	for i, h := range txs.Txs {
		str[i] = h.String()
	}
	return str
}

type TxIndex interface {
	GetAccountTxs(account [20]byte) (AccountTxs, error)
	GetTx(hash chainhash.Hash) (TxLocator, error)
	SetTx(locator TxLocator, account [20]byte) error
}

var _ TxIndex = &txIndex{}

// txIndex is a pseudo chainindex that contains locators for account transactions.
type txIndex struct {
	db *bbolt.DB
}

var accBucketPrefix = []byte("acc-")

// GetAccountTxs returns a list of transaction related to an account
func (i *txIndex) GetAccountTxs(account [20]byte) (AccountTxs, error) {
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
			kb := [32]byte{}
			copy(kb[:], k)
			txs.Amount++
			h, err := chainhash.NewHash(kb[:])
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

// GetTx returns a transaction locator from the chainindex.
func (i *txIndex) GetTx(hash chainhash.Hash) (TxLocator, error) {
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
func (i *txIndex) SetTx(locator TxLocator, account [20]byte) error {
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
		err = accBkt.Put(locator.Hash[:], []byte{})
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

// NewTxIndex returns/creates a new tx chainindex database
func NewTxIndex(datadir string) (TxIndex, error) {
	db, err := bbolt.Open(path.Join(datadir, "tx.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	return &txIndex{db: db}, nil
}
