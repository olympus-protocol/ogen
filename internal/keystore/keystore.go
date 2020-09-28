package keystore

import (
	"errors"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"go.etcd.io/bbolt"
	"path"
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
	AddKey(priv []byte) error
}

// keystore is a wrapper for the keystore database
type keystore struct {
	// db is a reference to the bbolt database
	db *bbolt.DB
	// datadir is the folder where the database is located
	datadir string
	// open prevents accessing the database when is closed
	open bool
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

func (k *keystore) load(db *bbolt.DB) error {
	err := db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(keysBucket)
		if bkt == nil {
			return ErrorNotInitialized
		}
		return nil
	})
	if err != nil {
		return err
	}
	k.db = db
	k.open = true
	return nil
}

// Close closes the keystore database
func (k *keystore) Close() error {
	k.open = false
	return k.db.Close()
}

// NewKeystore creates a new keystore instance.
func NewKeystore(datadir string) Keystore {
	return &keystore{
		open:    false,
		datadir: datadir,
	}
}
