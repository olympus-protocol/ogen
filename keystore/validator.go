package keystore

import (
	"github.com/olympus-protocol/ogen/bls"
	"go.etcd.io/bbolt"
)

var keysBucket = []byte("keys")

// GetValidatorKeys gets all keys.
func (k *Keystore) GetValidatorKeys() ([]*bls.SecretKey, error) {
	secKeys := make([]*bls.SecretKey, 0)
	err := k.db.View(func(txn *bbolt.Tx) error {
		bkt := txn.Bucket(keysBucket)
		bkt = txn.Bucket(keysBucket)
		err := bkt.ForEach(func(k, v []byte) error {
			var valBytes [32]byte
			copy(valBytes[:], v)
			secretKey, err := bls.SecretKeyFromBytes(valBytes[:])
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
func (k *Keystore) GenerateNewValidatorKey(amount uint64) ([]*bls.SecretKey, error) {
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
