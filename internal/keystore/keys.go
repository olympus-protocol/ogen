package keystore

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	"go.etcd.io/bbolt"
)

// GetValidatorKey returns the private key from the specified public key or false if doesn't exists.
func (k *keystore) GetValidatorKey(pubkey [48]byte) (*bls.SecretKey, bool) {

	if !k.open {
		return nil, false
	}

	var key []byte
	err := k.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(keysBucket)
		key = bkt.Get(pubkey[:])
		return nil
	})
	if err != nil {
		return nil, false
	}

	if key == nil {
		return nil, false
	}

	blsKey, err := bls.SecretKeyFromBytes(key)
	if err != nil {
		return nil, false
	}

	return blsKey, true
}

// GetValidatorKeys returns all keys on keystore.
func (k *keystore) GetValidatorKeys() ([]*bls.SecretKey, error) {

	if !k.open {
		return nil, ErrorNoOpen
	}

	var keys []*bls.SecretKey

	err := k.db.View(func(tx *bbolt.Tx) error {

		bkt := tx.Bucket(keysBucket)

		err := bkt.ForEach(func(keypub, keyprv []byte) error {

			key, err := bls.SecretKeyFromBytes(keyprv)
			if err != nil {
				return err
			}

			keys = append(keys, key)

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
	return keys, nil
}

// GenerateNewValidatorKey generates new validator keys and adds it to the map and database.
func (k *keystore) GenerateNewValidatorKey(amount uint64) ([]*bls.SecretKey, error) {
	if !k.open {
		return nil, ErrorNoOpen
	}

	keys := make([]*bls.SecretKey, amount)

	for i := range keys {
		// Generate a new key
		key := bls.RandKey()
		err := k.addKey(key)
		if err != nil {
			return nil, err
		}
		keys[i] = key
	}

	return keys, nil
}

func (k *keystore) AddKey(priv []byte) error {
	s, err := bls.SecretKeyFromBytes(priv)
	if err != nil {
		return err
	}
	return k.addKey(s)
}

func (k *keystore) addKey(priv *bls.SecretKey) error {

	if !k.open {
		return ErrorNoOpen
	}

	return k.db.Update(func(tx *bbolt.Tx) error {

		bkt := tx.Bucket(keysBucket)

		err := bkt.Put(priv.PublicKey().Marshal(), priv.Marshal())
		if err != nil {
			return err
		}

		return nil
	})
}
