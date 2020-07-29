package index

import (
	"path"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"go.etcd.io/bbolt"
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
			kb := [32]byte{}
			copy(kb[:], k)
			txs.Amount++
			h, err := chainhash.NewHash(kb)
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
		locHash := locator.Hash.CloneBytes()
		err = accBkt.Put(locHash[:], []byte{})
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
	Hash  [32]byte
	Block [32]byte
	Index uint32
}

// Marshal encodes the data.
func (tl *TxLocator) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(tl)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (tl *TxLocator) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(d, tl)
}

// NewTxIndex returns/creates a new tx index database
func NewTxIndex(datadir string) (*TxIndex, error) {
	db, err := bbolt.Open(path.Join(datadir, "tx.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	return &TxIndex{db: db}, nil
}
