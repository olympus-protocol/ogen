package keystore

import (
	"errors"
	"path"
	"sync"

	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"go.etcd.io/bbolt"
)

var (

	// ErrorNotInitialized is returned when a bucket is not properly initialized
	ErrorNotInitialized = errors.New("the keystore is not initialized")

	// ErrorAlreadyOpen returned when user tries to open a keystore already open.
	ErrorAlreadyOpen = errors.New("keystore is already open")

	// ErrorNoOpen returned when the keystore is not open.
	ErrorNoOpen = errors.New("open the keystore to start using it")

	// ErrorKeystoreExists returned when a user creates a new keystore over an existing keystore.
	ErrorKeystoreExists = errors.New("cannot create new keystore, it already exists")
)

var (
	// keysBucket is the bucket key for the keystore keys
	keysBucket = []byte("keys")
)

type Keystore interface {
	CreateKeystore() error
	OpenKeystore() error
	Close() error
	GetValidatorKey(pubkey [48]byte) (*bls.SecretKey, bool)
	GetValidatorKeys() ([]*bls.SecretKey, error)
	GenerateNewValidatorKey(amount uint64) ([]*bls.SecretKey, error)
	HasKeysToParticipate() bool
}

// keystore is a wrapper for the keystore database
type keystore struct {
	// db is a reference to the bbolt database
	db *bbolt.DB
	// log is a reference to the global logger
	log logger.Logger
	// datadir is the folder where the database is located
	datadir string
	// open prevents accessing the database when is closed
	open bool
	// keys is a memory map for access the keys faster
	keys map[chainhash.Hash]*bls.SecretKey
	// keysLock is the maps lock
	keysLock sync.RWMutex
}

var _ Keystore = &keystore{}

func (k *keystore) HasKeysToParticipate() bool {
	keys, err := k.GetValidatorKeys()
	if err != nil {
		return false
	}
	return len(keys) > 0
}

// CreateKeystore will create a new keystore and initialize it.
func (k *keystore) CreateKeystore() error {
	if k.open {
		return ErrorAlreadyOpen
	}
	// Open the database
	db, err := bbolt.Open(path.Join(k.datadir, "keystore.db"), 0600, nil)
	if err != nil {
		return err
	}
	err = k.initialize(db)
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
func (k *keystore) OpenKeystore() error {
	if k.open {
		return ErrorAlreadyOpen
	}
	// Open the database
	db, err := bbolt.Open(path.Join(k.datadir, "keystore.db"), 0600, nil)
	if err != nil {
		return err
	}
	err = k.load(db)
	if err != nil {
		_ = db.Close()
		return err
	}
	return nil
}

// load is used to properly fill the Keystore data with the database information.
// returns ErrorNotInitialized if the database doesn't exists or is not properly initialized.
func (k *keystore) load(db *bbolt.DB) error {

	keysMap := make(map[chainhash.Hash]*bls.SecretKey)

	// Load the keys to a memory map
	err := db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(keysBucket)
		if bkt == nil {
			return ErrorNotInitialized
		}

		err := bkt.ForEach(func(keypub, keyprv []byte) error {

			pubHash := chainhash.HashH(keypub)

			key, err := bls.SecretKeyFromBytes(keyprv)
			if err != nil {
				return err
			}

			keysMap[pubHash] = key

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
	k.db = db
	k.keys = keysMap
	k.open = true

	return nil
}

func (k *keystore) initialize(db *bbolt.DB) error {

	err := db.Update(func(tx *bbolt.Tx) error {

		_, err := tx.CreateBucket(keysBucket)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return k.load(db)
}

// Close closes the keystore database
func (k *keystore) Close() error {
	k.open = false
	k.keys = map[chainhash.Hash]*bls.SecretKey{}
	return k.db.Close()
}

// NewKeystore creates a new keystore instance.
func NewKeystore(datadir string, log logger.Logger) Keystore {
	return &keystore{
		log:     log,
		open:    false,
		datadir: datadir,
	}
}
