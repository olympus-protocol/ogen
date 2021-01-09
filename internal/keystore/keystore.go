package keystore

import (
	"encoding/binary"
	"errors"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/bip39"
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

	// ErrorKeyNotOnKeystore returned when tried to fetch a key that is not on the keystore
	ErrorKeyNotOnKeystore = errors.New("the specified public key doesn't exists on the keystore")
)

var (
	// keysBucket is the bucket key for the keystore keys
	keysBucket     = []byte("keys")
	mnemonicBucket = []byte("mnemonic")
	mnemonicKey    = []byte("mnemonic-key")
	lastPathBkt    = []byte("last-path")
	lastPathKey    = []byte("last-path-key")
)

type Keystore interface {
	CreateKeystore() error
	OpenKeystore() error
	Close() error
	GenerateNewValidatorKey(amount uint64) ([]*Key, error)
	HasKeysToParticipate() bool

	GetValidatorKey(pubkey [48]byte) (*Key, bool)
	GetValidatorKeys() ([]*Key, error)
	GetMnemonic() string
	GetLastPath() int

	ToggleKey(pub [48]byte, value bool) error
	AddKey(k *Key) error
}

// keystore is a wrapper for the keystore database
type keystore struct {
	// db is a reference to the bbolt database
	db *bbolt.DB
	// mnemonic is the mnemonic key used to derive keys
	mnemonic string
	// lastPath is the last used path from the keys derived
	lastPath int
	// datadir is the folder where the database is located
	datapath string
	// open prevents accessing the database when is closed
	open bool
}

var _ Keystore = &keystore{}

// CreateKeystore will create a new keystore and initialize it with a new mnemonic.
func (k *keystore) CreateKeystore() error {
	if k.open {
		return ErrorAlreadyOpen
	}
	// Open the database
	db, err := bbolt.Open(path.Join(k.datapath, "keystore.db"), 0600, nil)
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
	db, err := bbolt.Open(path.Join(k.datapath, "keystore.db"), 0600, nil)
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

// Close closes the keystore database
func (k *keystore) Close() error {
	k.open = false
	return k.db.Close()
}

// GetMnemonic returns the keystore mnemonic string
func (k *keystore) GetMnemonic() string {
	return k.mnemonic
}

// GetLastPath returns the last used path for key derivation
func (k *keystore) GetLastPath() int {
	return k.lastPath
}

// SetLastPath modifies the keystore last path
func (k *keystore) SetLastPath(p int) error {
	err := k.db.Update(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(lastPathBkt)
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(p))

		err := bkt.Put(lastPathKey, buf[:])
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	k.lastPath = p

	return nil
}

func (k *keystore) initialize(db *bbolt.DB) error {

	err := db.Update(func(tx *bbolt.Tx) error {

		_, err := tx.CreateBucketIfNotExists(keysBucket)
		if err != nil {
			return err
		}

		mnemonicBkt, err := tx.CreateBucketIfNotExists(mnemonicBucket)
		if err != nil {
			return err
		}

		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			return err
		}

		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			return err
		}

		err = mnemonicBkt.Put(mnemonicKey, []byte(mnemonic))
		if err != nil {
			return err
		}

		lastPathBkt, err := tx.CreateBucketIfNotExists(lastPathBkt)
		if err != nil {
			return err
		}

		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], 0)
		err = lastPathBkt.Put(lastPathKey, buf[:])
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
	var mnemonicBytes []byte
	err = db.View(func(tx *bbolt.Tx) error {
		mnemonic := tx.Bucket(mnemonicBucket)
		if mnemonic == nil {
			return ErrorNotInitialized
		}
		mnemonicBytes = mnemonic.Get(mnemonicKey)
		return nil
	})

	var lastPathBytes []byte
	err = db.View(func(tx *bbolt.Tx) error {
		lastPath := tx.Bucket(lastPathBkt)
		if lastPath == nil {
			return ErrorNotInitialized
		}
		lastPathBytes = lastPath.Get(lastPathKey)
		return nil
	})

	if err != nil {
		return err
	}
	k.db = db
	k.open = true
	k.mnemonic = string(mnemonicBytes)
	k.lastPath = int(binary.LittleEndian.Uint64(lastPathBytes))
	return nil
}

// NewKeystore creates a new keystore instance.
func NewKeystore() Keystore {
	return &keystore{
		open:     false,
		datapath: config.GlobalFlags.DataPath,
	}
}
