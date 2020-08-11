package keystore

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"go.etcd.io/bbolt"
)

// GetValidatorKey returns the private key from the specified public key or false if doesn't exists.
func (k *Keystore) GetValidatorKey(pubkey [48]byte) (*bls.SecretKey, bool) {

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
func (k *Keystore) GetValidatorKeys() ([]*bls.SecretKey, error) {

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
func (k *Keystore) GenerateNewValidatorKey(amount uint64) ([]*bls.SecretKey, error) {
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

func (k *Keystore) addKey(priv *bls.SecretKey) error {

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

func (k *Keystore) addKeyMap(hash chainhash.Hash, key *bls.SecretKey) error {

	if !k.open {
		return ErrorNoOpen
	}

	k.keysLock.Lock()
	defer k.keysLock.Unlock()

	k.keys[hash] = key
	return nil
}

func (k *Keystore) addKeyDB(encryptedKey []byte, pubkey []byte) error {

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
