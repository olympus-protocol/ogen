package keystore

import (
	"path"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/logger"
	"go.etcd.io/bbolt"
)

var keysBucket = []byte("keys")

// Keystore is a wrapper for the keystore database
type Keystore struct {
	db  *bbolt.DB
	log *logger.Logger
}

// NewKeystore opens a keystore or creates a new one if doesn't exist.
func NewKeystore(pathStr string, log *logger.Logger) (*Keystore, error) {
	db, err := bbolt.Open(path.Join(pathStr, "keystore.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(keysBucket); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	w := &Keystore{
		db:  db,
		log: log,
	}
	return w, nil
}

// Close closes the keystore database
func (k *Keystore) Close() error {
	return k.db.Close()
}

// GetValidatorKeys gets all keys.
func (k *Keystore) GetValidatorKeys() ([]*bls.SecretKey, error) {
	secKeys := []*bls.SecretKey{}
	err := k.db.View(func(txn *bbolt.Tx) error {
		bkt := txn.Bucket(keysBucket)
		err := bkt.ForEach(func(k, v []byte) error {
			secretKey, err := bls.SecretKeyFromBytes(v)
			if err != nil {
				return err
			}
			secKeys = append(secKeys, secretKey)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return secKeys, nil
}

// GetValidatorKey gets a validator key from the key store.
func (k *Keystore) GetValidatorKey(pubkey []byte) (*bls.SecretKey, bool) {
	var secretBytes []byte
	err := k.db.View(func(txn *bbolt.Tx) error {
		bkt := txn.Bucket(keysBucket)
		secretBytes = bkt.Get(pubkey)
		return nil
	})
	if err != nil {
		return nil, false
	}
	secretKey, err := bls.SecretKeyFromBytes(secretBytes)
	if err != nil {
		return nil, false
	}
	return secretKey, true
}

// HasValidatorKey checks if a validator key exists.
func (k *Keystore) HasValidatorKey(pubBytes []byte) (result bool, err error) {
	err = k.db.View(func(txn *bbolt.Tx) error {
		bkt := txn.Bucket(keysBucket)
		key := bkt.Get(pubBytes)
		if key == nil {
			result = false
		} else {
			result = true
		}
		return nil
	})

	return result, err
}

// GenerateNewValidatorKey generates a new validator key.
func (k *Keystore) GenerateNewValidatorKey(amount uint64, password string) ([]*bls.SecretKey, error) {
	keys := make([]*bls.SecretKey, amount)
	for i := range keys {
		key := bls.RandKey()
		keyBytes := key.Marshal()
		pub := key.PublicKey()
		pubBytes := pub.Marshal()
		err := k.db.Update(func(txn *bbolt.Tx) error {
			bkt := txn.Bucket(keysBucket)
			return bkt.Put(pubBytes, keyBytes)
		})
		if err != nil {
			return nil, err
		}
		keys[i] = key
	}
	return keys, nil
}
