package keystore

import (
	"errors"
	"path"
	"reflect"
	"sync"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/blsaes"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/logger"
	"go.etcd.io/bbolt"
)

var (
	passHashKey = []byte("passhash")
	saltKey = []byte("salt")
	nonceKey = []byte("nonce")
	infoBucket = []byte("info")
	keysBucket = []byte("keys")
)

// Keystore is a wrapper for the keystore database
type Keystore struct {
	db  *bbolt.DB
	log *logger.Logger
	open bool
	keys map[chainhash.Hash]*bls.SecretKey
	keysLock sync.RWMutex
}

// AddKey is a public function to add a new key to the map and to the database.
func (k *Keystore) AddKey(hash chainhash.Hash, key *bls.SecretKey) {
	k.keysLock.Lock()
	k.keys[hash] = key
	k.keysLock.Unlock()
	return
}

func (k *Keystore) addKeyDB() {
	// TODO
}

// NewKeystore opens a keystore or creates a new one if doesn't exist.
func NewKeystore(pathStr string, log *logger.Logger, password string) (*Keystore, error) {
	// Open the database
	db, err := bbolt.Open(path.Join(pathStr, "keystore.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	w := &Keystore{
		db:  db,
		log: log,
		open: true,
	}
	// Open the keystore and check password hashes.
	err = w.load(password)
	if err != nil {
		err = w.initialize(password)
		if err != nil {
			return nil, err
		}
	}
	
	return w, nil
}

func (k *Keystore) load(password string) error {
	err := k.db.View(func( tx *bbolt.Tx) error {
		infobkt := tx.Bucket(infoBucket)
		if infobkt == nil {
			return errors.New("not initialized")
		}
		currPassHash := chainhash.HashB([]byte(password))
		passhash := infobkt.Get(passHashKey)
		if equal := reflect.DeepEqual(passhash, currPassHash); !equal {
			return errors.New("password don't match")
		}
		var salt [8]byte
		var nonce [12]byte
		saltB := infobkt.Get(saltKey)
		nonceB := infobkt.Get(nonceKey)
		copy(salt[:], saltB)
		copy(nonce[:], nonceB)
		keysbkt := tx.Bucket(keysBucket)
		if keysbkt == nil {
			return errors.New("not initialized")
		}
		k.keys = make(map[chainhash.Hash]*bls.SecretKey)
		keysbkt.ForEach(func(key, v []byte) error {
			pubHash := chainhash.HashH(key)
			priv, err := blsaes.Decrypt(v, nonce, []byte(password), salt)
			if err != nil {
				return err
			}
			k.AddKey(pubHash, priv)
			return nil
		})
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (k *Keystore) initialize(password string) error {
	return nil
}

// Close closes the keystore database
func (k *Keystore) Close() error {
	k.open = false
	k.keys = map[chainhash.Hash]*bls.SecretKey{}
	return k.db.Close()
}

// GetValidatorKeys gets all keys.
func (k *Keystore) GetValidatorKeys() ([]*bls.SecretKey, error) {
	keys := []*bls.SecretKey{}
	for _, k := range keys {
		keys = append(keys, k)
	}
	return keys, nil
}

// GetValidatorKey gets a validator key from the key store.
func (k *Keystore) GetValidatorKey(pubkey []byte) (*bls.SecretKey, bool) {
	pubHash := chainhash.HashH(pubkey)
	k.keysLock.Lock()
	key, ok := k.keys[pubHash]
	k.keysLock.Unlock()
	return key, ok
}

// HasValidatorKey checks if a validator key exists.
func (k *Keystore) HasValidatorKey(pubkey []byte) bool {
	pubHash := chainhash.HashH(pubkey)
	k.keysLock.Lock()
	_, ok := k.keys[pubHash]
	k.keysLock.Unlock()
	return ok
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
