package keystore

import (
	"github.com/olympus-protocol/ogen/bls"
	"go.etcd.io/bbolt"
)

var keysBucket = []byte("keys")

func (k *Keystore) GetValidatorKeys() ([]*bls.SecretKey, error) {
	secKeys := make([]*bls.SecretKey, 0)
	err := k.db.Update(func(txn *bbolt.Tx) error {
		bkt, err := txn.CreateBucketIfNotExists(keysBucket)
		if err != nil {
			return err
		}
		err = bkt.ForEach(func(k, v []byte) error {
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

func (k *Keystore) GetValidatorKey(pubkey []byte) (*bls.SecretKey, bool) {
	var secretBytes []byte
	err := k.db.Update(func(txn *bbolt.Tx) error {
		bkt, err := txn.CreateBucketIfNotExists(keysBucket)
		if err != nil {
			return err
		}
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

func (k *Keystore) HasValidatorKey(pubBytes []byte) (result bool, err error) {
	err = k.db.View(func(txn *bbolt.Tx) error {
		bkt, err := txn.CreateBucketIfNotExists(keysBucket)
		if err != nil {
			return err
		}
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

func (k *Keystore) GenerateNewValidatorKey() (*bls.SecretKey, error) {
	key := bls.RandKey()
	keyBytes := key.Marshal()
	pub := key.PublicKey()
	pubBytes := pub.Marshal()
	err := k.db.Update(func(txn *bbolt.Tx) error {
		bkt, err := txn.CreateBucketIfNotExists(keysBucket)
		if err != nil {
			return err
		}
		return bkt.Put(pubBytes, keyBytes)
	})
	return key, err
}
