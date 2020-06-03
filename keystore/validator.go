package keystore

import (
	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/bls"
)

func (k *Keystore) GetValidatorKeys() ([]*bls.SecretKey, error) {
	secKeys := make([]*bls.SecretKey, 0)
	err := k.db.View(func(txn *badger.Txn) error {
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
				secretKey, err := bls.SecretKeyFromBytes(valBytes[:])
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

func (k *Keystore) GetValidatorKey(pubkey []byte) (*bls.SecretKey, bool) {
	var secretBytes []byte
	err := k.db.View(func(txn *badger.Txn) error {
		i, err := txn.Get(pubkey)
		if err != nil {
			return err
		}

		_, err = i.ValueCopy(secretBytes)
		return err
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
	err = k.db.View(func(txn *badger.Txn) error {
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

func (k *Keystore) GenerateNewValidatorKey() (*bls.SecretKey, error) {
	key := bls.RandKey()
	keyBytes := key.Marshal()
	pub := key.PublicKey()
	pubBytes := pub.Marshal()
	err := k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(pubBytes, keyBytes)
	})
	return key, err
}
