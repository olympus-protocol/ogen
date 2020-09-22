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

	pub, err := bls.PublicKeyFromBytes(pubkey[:])
	if err != nil {
		return nil, false
	}

	// TODO find a more optimized way to ensure keysMap and keysDB match.

	keysCount, err := k.getDBKeysCount()
	if err != nil {
		return nil, false
	}

	if keysCount != len(k.keys) {
		err = k.reloadKeysMap()
		if err != nil {
			return nil, false
		}
	}

	k.keysLock.Lock()
	defer k.keysLock.Unlock()

	key, ok := k.keys[pub]

	return key, ok
}

// GetValidatorKeys returns all keys on keystore.
func (k *keystore) GetValidatorKeys() ([]*bls.SecretKey, error) {

	if !k.open {
		return nil, ErrorNoOpen
	}

	defer k.keysLock.Unlock()

	k.keysLock.Lock()

	var keys []*bls.SecretKey

	for _, k := range k.keys {
		keys = append(keys, k)
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

	err := k.addKeyDB(priv)
	if err != nil {
		return err
	}

	err = k.addKeyMap(priv)
	if err != nil {
		return err
	}

	return nil
}

func (k *keystore) addKeyMap(key *bls.SecretKey) error {

	if !k.open {
		return ErrorNoOpen
	}

	k.keysLock.Lock()
	defer k.keysLock.Unlock()

	k.keys[key.PublicKey()] = key
	return nil
}

func (k *keystore) addKeyDB(key *bls.SecretKey) error {

	if !k.open {
		return ErrorNoOpen
	}

	return k.db.Update(func(tx *bbolt.Tx) error {

		bkt := tx.Bucket(keysBucket)

		err := bkt.Put(key.PublicKey().Marshal(), key.Marshal())
		if err != nil {
			return err
		}

		return nil
	})

}

func (k *keystore) getDBKeysCount() (int, error) {
	if !k.open {
		return 0, ErrorNoOpen
	}
	count := 0
	err := k.db.View(func(tx *bbolt.Tx) error {

		bkt := tx.Bucket(keysBucket)

		stats := bkt.Stats()
		count = stats.KeyN

		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (k *keystore) reloadKeysMap() error {
	k.keysLock.Lock()
	defer k.keysLock.Unlock()

	k.keys = make(map[*bls.PublicKey]*bls.SecretKey)

	return k.db.Update(func(tx *bbolt.Tx) error {

		bkt := tx.Bucket(keysBucket)

		err := bkt.ForEach(func(keypub, keyprv []byte) error {

			key, err := bls.SecretKeyFromBytes(keyprv)
			if err != nil {
				return err
			}

			k.keys[key.PublicKey()] = key

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
}