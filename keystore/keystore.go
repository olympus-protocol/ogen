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

	// ErrorNotInitialized is returned when a bucket is not properly initialized
	ErrorNotInitialized = errors.New("the keystore is not initialized")

	// ErrorLoadingEncryptionData is returned when the encryption data is null
	ErrorLoadingEncryptionData = errors.New("encryption information is empty")

	// ErrorAlreadyOpen returned when user tries to open a keystore already open.
	ErrorAlreadyOpen = errors.New("keystore is already open")

	// ErrorNoOpen returned when the keystore is not open.
	ErrorNoOpen = errors.New("open the keystore to start using it")

	// ErrorKeystoreExists returned when a user creates a new keystore over an existing keystore.
	ErrorKeystoreExists = errors.New("cannot create new keystore, it already exists")

	// ErrorPassNoMatch is returned when the keystore password don't match.
	ErrorPassNoMatch = errors.New("the provided password doesn't match with current keystore password")
)

var (
	// passHashKey is the key of the password hash
	passHashKey = []byte("passhash")
	// saltKey is the key of the encryption salt
	saltKey = []byte("salt")
	// nonceKey is the key of the encryption nonce
	nonceKey = []byte("nonce")
	// infoBucket is the bucket key for the keystore information
	infoBucket = []byte("info")
	// keysBucket is the bucket key for the keystore keys
	keysBucket = []byte("keys")
)

type encryptionInfo struct {
	nonce    [12]byte
	salt     [8]byte
	passhash []byte
}

// Keystore is a wrapper for the keystore database
type Keystore struct {
	// db is a reference to the bbolt database
	db *bbolt.DB
	// log is a reference to the global logger
	log *logger.Logger
	// datadir is the folder where the database is located
	datadir string
	// open prevents accessing the database when is closed
	open bool
	// keys is a memory map for access the keys faster
	keys map[chainhash.Hash]*bls.SecretKey
	// keysLock is the maps lock
	keysLock sync.RWMutex
	// encryptionInfo is the current loaded database encryption information
	encryptionInfo encryptionInfo
}

// CreateKeystore will create a new keystore and initialize it.
func (k *Keystore) CreateKeystore(password string) error {
	if k.open {
		return ErrorAlreadyOpen
	}
	// Open the database
	db, err := bbolt.Open(path.Join(k.datadir, "keystore.db"), 0600, nil)
	if err != nil {
		return err
	}
	err = k.initialize(password, db)
	if err != nil {
		_ = db.Close()
		if err == bbolt.ErrBucketExists {
			return ErrorKeystoreExists
		}
		return err
	}
	return nil
}

// OpenKeystore opens an already created keystore, returns error if the Keystore is previously initialized.
func (k *Keystore) OpenKeystore(password string) error {
	if k.open {
		return ErrorAlreadyOpen
	}
	// Open the database
	db, err := bbolt.Open(path.Join(k.datadir, "keystore.db"), 0600, nil)
	if err != nil {
		return err
	}
	err = k.load(password, db)
	if err != nil {
		_ = db.Close()
		return err
	}
	return nil
}

// load is used to properly fill the Keystore data with the database information.
// returns ErrorNotInitialized if the database doesn't exists or is not properly initialized.
func (k *Keystore) load(password string, db *bbolt.DB) error {

	var nonce [12]byte
	var salt [8]byte
	var passhash []byte

	// Load the encryption data
	err := db.View(func(tx *bbolt.Tx) error {

		bkt := tx.Bucket(infoBucket)
		if bkt == nil {
			return ErrorNotInitialized
		}

		saltB := bkt.Get(saltKey)
		nonceB := bkt.Get(nonceKey)
		passhash = bkt.Get(passHashKey)

		copy(salt[:], saltB)
		copy(nonce[:], nonceB)

		return nil
	})
	if err != nil {
		return err
	}

	// Check not null values on encryption data
	if salt == [8]byte{} ||
		nonce == [12]byte{} ||
		passhash == nil {
		return ErrorLoadingEncryptionData
	}

	// Verify that password provided matches password hash
	equal := reflect.DeepEqual(passhash, chainhash.HashB([]byte(password)))
	if !equal {
		return ErrorPassNoMatch
	}

	keysMap := make(map[chainhash.Hash]*bls.SecretKey)

	// Load the keys to a memory map
	err = db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(keysBucket)
		if bkt == nil {
			return ErrorNotInitialized
		}

		err = bkt.ForEach(func(keypub, keyprv []byte) error {

			pubHash := chainhash.HashH(keypub)

			priv, err := aesbls.Decrypt(nonce, salt, keyprv, []byte(password))
			if err != nil {
				return err
			}

			keysMap[pubHash] = priv

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

	// If everything goes correctly, we add the elements to the struct
	k.encryptionInfo = encryptionInfo{
		nonce:    nonce,
		salt:     salt,
		passhash: passhash,
	}
	k.db = db
	k.keys = keysMap
	k.open = true

	return nil
}

func (k *Keystore) initialize(password string, db *bbolt.DB) error {

	salt, err := aesbls.RandSalt()
	if err != nil {
		return err
	}

	nonce, err := aesbls.RandNonce()
	if err != nil {
		return err
	}

	passhash := chainhash.HashB([]byte(password))

	err = db.Update(func(tx *bbolt.Tx) error {

		bkt, err := tx.CreateBucket(infoBucket)
		if err != nil {
			return err
		}

		err = bkt.Put(saltKey, salt[:])
		if err != nil {
			return err
		}

		err = bkt.Put(nonceKey, nonce[:])
		if err != nil {
			return err
		}

		err = bkt.Put(passHashKey, passhash[:])
		if err != nil {
			return err
		}

		_, err = tx.CreateBucket(keysBucket)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return k.load(password, db)
}

func (k *Keystore) checkPassword(password string) error {

	if !k.open {
		return ErrorNoOpen
	}

	currPassHash := chainhash.HashB([]byte(password))
	if equal := reflect.DeepEqual(k.encryptionInfo.passhash, currPassHash); !equal {
		return ErrorPassNoMatch
	}

	return nil
}

// Close closes the keystore database
func (k *Keystore) Close() error {
	k.open = false
	k.keys = map[chainhash.Hash]*bls.SecretKey{}
	return k.db.Close()
}

// NewKeystore creates a new keystore instance.
func NewKeystore(datadir string, log *logger.Logger) *Keystore {
	return &Keystore{
		log:     log,
		open:    false,
		datadir: datadir,
	}
}
