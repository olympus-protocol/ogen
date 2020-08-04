package keystore

import (
	"errors"
	"path"
	"reflect"
	"sync"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/aesbls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/logger"
	"go.etcd.io/bbolt"
)

var (
	errorNotInitialized = errors.New("not initialized")

	passHashKey = []byte("passhash")
	saltKey     = []byte("salt")
	nonceKey    = []byte("nonce")
	infoBucket  = []byte("info")
	keysBucket  = []byte("keys")
)

// Keystore is a wrapper for the keystore database
type Keystore struct {
	db       *bbolt.DB
	log      *logger.Logger
	open     bool
	keys     map[chainhash.Hash]*bls.SecretKey
	keysLock sync.RWMutex
}

func (k *Keystore) addKey(priv *bls.SecretKey, password string) error {
	if !k.open {
		return errors.New("keystore not open, please open it first")
	}
	nonce, salt, err := k.getEncryptionData()
	if err != nil {
		return err
	}
	encryptedKey, err := aesbls.SimpleEncrypt(priv.Marshal(), []byte(password), nonce, salt)
	if err != nil {
		return err
	}
	err = k.addKeyDB(encryptedKey, priv.PublicKey().Marshal())
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
		return errors.New("keystore not open, please open it first")
	}
	k.keysLock.Lock()
	k.keys[hash] = key
	k.keysLock.Unlock()
	return nil
}

func (k *Keystore) addKeyDB(encryptedKey []byte, pubkey []byte) error {
	return k.db.Update(func(tx *bbolt.Tx) error {
		keysbkt := tx.Bucket(keysBucket)
		err := keysbkt.Put(pubkey, encryptedKey)
		if err != nil {
			return err
		}
		return nil
	})
}

// NewKeystore opens a keystore or creates a new one if doesn't exist.
func NewKeystore(pathStr string, log *logger.Logger, password string) (*Keystore, error) {
	// Open the database
	db, err := bbolt.Open(path.Join(pathStr, "keystore.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	w := &Keystore{
		db:   db,
		log:  log,
		open: true,
	}
	// Open the keystore and check password hashes.
load:
	err = w.load(password)
	if err != nil {
		if err == errorNotInitialized {
			err = w.initialize(password)
			if err != nil {
				return nil, err
			}
			goto load
		}
		return nil, err
	}

	return w, nil
}

func (k *Keystore) load(password string) error {

	valid, err := k.checkPassword(password)
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("invalid password")
	}
	nonce, salt, err := k.getEncryptionData()

	if err != nil {
		return err
	}
	err = k.db.View(func(tx *bbolt.Tx) error {

		keysbkt := tx.Bucket(keysBucket)
		if keysbkt == nil {
			return errorNotInitialized
		}

		k.keys = make(map[chainhash.Hash]*bls.SecretKey)

		err = keysbkt.ForEach(func(keypub, keyprv []byte) error {

			pubHash := chainhash.HashH(keypub)
			priv, err := aesbls.Decrypt(nonce, salt, keyprv, []byte(password))
			if err != nil {
				return err
			}

			err = k.addKeyMap(pubHash, priv)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (k *Keystore) initialize(password string) error {
	salt, err := aesbls.RandSalt()
	if err != nil {
		return err
	}
	nonce, err := aesbls.RandNonce()
	if err != nil {
		return err
	}
	passhash := chainhash.HashB([]byte(password))
	return k.db.Update(func(tx *bbolt.Tx) error {
		infobkt, err := tx.CreateBucket(infoBucket)
		if err != nil {
			return err
		}
		err = infobkt.Put(saltKey, salt[:])
		if err != nil {
			return err
		}
		err = infobkt.Put(nonceKey, nonce[:])
		if err != nil {
			return err
		}
		err = infobkt.Put(passHashKey, passhash[:])
		if err != nil {
			return err
		}
		_, err = tx.CreateBucket(keysBucket)
		if err != nil {
			return err
		}
		return nil
	})
}

func (k *Keystore) getEncryptionData() (nonce [12]byte, salt [8]byte, err error) {
	err = k.db.View(func(tx *bbolt.Tx) error {
		infobkt := tx.Bucket(infoBucket)
		if infobkt == nil {
			return errorNotInitialized
		}
		saltB := infobkt.Get(saltKey)
		nonceB := infobkt.Get(nonceKey)
		copy(salt[:], saltB)
		copy(nonce[:], nonceB)
		return nil
	})
	if err != nil {
		return [12]byte{}, [8]byte{}, err
	}
	return nonce, salt, nil
}

func (k *Keystore) checkPassword(password string) (bool, error) {
	if !k.open {
		return false, nil
	}
	err := k.db.View(func(tx *bbolt.Tx) error {
		infobkt := tx.Bucket(infoBucket)
		if infobkt == nil {
			return errorNotInitialized
		}
		currPassHash := chainhash.HashB([]byte(password))
		passhash := infobkt.Get(passHashKey)
		if equal := reflect.DeepEqual(passhash, currPassHash); !equal {
			return errors.New("password don't match")
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// Close closes the keystore database
func (k *Keystore) Close() error {
	k.open = false
	k.keys = map[chainhash.Hash]*bls.SecretKey{}
	return k.db.Close()
}

// GetValidatorKeys gets all keys.
func (k *Keystore) GetValidatorKeys() ([]*bls.SecretKey, error) {
	var keys []*bls.SecretKey
	k.keysLock.Lock()
	for _, k := range k.keys {
		keys = append(keys, k)
	}
	k.keysLock.Unlock()
	return keys, nil
}

// GetValidatorKey gets a validator key from the key store.
func (k *Keystore) GetValidatorKey(pubkey [48]byte) (*bls.SecretKey, bool) {
	pubHash := chainhash.HashH(pubkey[:])
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

// GenerateNewValidatorKey generates new validator keys and adds it to the map and database.
func (k *Keystore) GenerateNewValidatorKey(amount uint64, password string) ([]*bls.SecretKey, error) {
	keys := make([]*bls.SecretKey, amount)
	for i := range keys {
		// Generate a new key
		key := bls.RandKey()
		err := k.addKey(key, password)
		if err != nil {
			return nil, err
		}
		keys[i] = key
	}
	return keys, nil
}
