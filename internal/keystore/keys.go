package keystore

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"go.etcd.io/bbolt"
)

// GetValidatorKey returns the private key from the specified public key or false if doesn't exists.
func (k *keystore) GetValidatorKey(pubkey [48]byte) (*bls.SecretKey, bool) {

	if !k.open {
		return nil, false
	}

	k.keysLock.Lock()
	defer k.keysLock.Unlock()

	pubHash := chainhash.HashH(pubkey[:])

	key, ok := k.keys[pubHash]

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

	err := k.addKeyDB(priv.Marshal(), priv.PublicKey().Marshal())
	if err != nil {
		return err
	}

	err = k.addKeyMap(chainhash.HashH(priv.PublicKey().Marshal()), priv)
	if err != nil {
		return err
	}

	return nil
}

func (k *keystore) addKeyMap(hash chainhash.Hash, key *bls.SecretKey) error {

	if !k.open {
		return ErrorNoOpen
	}

	k.keysLock.Lock()
	defer k.keysLock.Unlock()

	k.keys[hash] = key
	return nil
}

func (k *keystore) addKeyDB(encryptedKey []byte, pubkey []byte) error {

	if !k.open {
		return ErrorNoOpen
	}

	return k.db.Update(func(tx *bbolt.Tx) error {

		bkt := tx.Bucket(keysBucket)

		err := bkt.Put(pubkey, encryptedKey)
		if err != nil {
			return err
		}

		return nil
	})
}
