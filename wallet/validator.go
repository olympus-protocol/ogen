package wallet

import (
	"crypto/rand"

	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/primitives"
)

type ValidatorWallet struct {
	db *badger.DB
}

// Keystore is an interface to a simple keystore.
type Keystore interface {
	GenerateNewValidatorKey() (*bls.SecretKey, error)
	GetValidatorKey(*primitives.Worker) (*bls.SecretKey, bool)
	HasValidatorKey([48]byte) (bool, error)
	GetValidatorKeys() ([]*bls.SecretKey, error)
	Close() error
}

func NewValidatorWallet(walletDB *badger.DB) *ValidatorWallet {
	return &ValidatorWallet{walletDB}
}

var _ Keystore = &ValidatorWallet{}

func (vw *ValidatorWallet) Close() error {
	return vw.db.Close()
}

func (vw *ValidatorWallet) GetValidatorKeys() ([]*bls.SecretKey, error) {
	secKeys := make([]*bls.SecretKey, 0)
	err := vw.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			i := iter.Item()
			val, err := i.ValueCopy(nil)
			if err != nil {
				return err
			}
			if len(val) == 32 {
				var valBytes [32]byte
				copy(valBytes[:], val)
				secretKey, err := bls.DeserializeSecretKey(valBytes)
				if err != nil {
					return err
				}
				secKeys = append(secKeys, secretKey)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return secKeys, nil
}

func (vw *ValidatorWallet) GetValidatorKey(worker *primitives.Worker) (*bls.SecretKey, bool) {
	pubBytes := worker.PubKey

	var secretBytes [32]byte
	err := vw.db.View(func(txn *badger.Txn) error {
		i, err := txn.Get(pubBytes[:])
		if err != nil {
			return err
		}

		_, err = i.ValueCopy(secretBytes[:])
		return err
	})
	if err != nil {
		return nil, false
	}

	secretKey, err := bls.DeserializeSecretKey(secretBytes)
	if err != nil {
		return nil, false
	}
	return secretKey, true
}

func (vw *ValidatorWallet) HasValidatorKey(pubBytes [48]byte) (result bool, err error) {
	err = vw.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(pubBytes[:])
		if err == badger.ErrKeyNotFound {
			result = false
		}
		if err != nil {
			return err
		}
		result = true
		return nil
	})

	return result, err
}

func (vw *ValidatorWallet) GenerateNewValidatorKey() (*bls.SecretKey, error) {
	key, err := bls.RandSecretKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	keyBytes := key.Serialize()

	pub := key.DerivePublicKey()
	pubBytes := pub.Serialize()

	err = vw.db.Update(func(txn *badger.Txn) error {
		return txn.Set(pubBytes[:], keyBytes[:])
	})

	return key, err
}
