package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/hdwallets"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/ssh/terminal"
)

var PolisNetPrefix = &hdwallets.NetPrefix{
	ExtPub:  []byte{0x1f, 0x74, 0x90, 0xf0},
	ExtPriv: []byte{0x11, 0x24, 0xd9, 0x70},
}

type Config struct {
	Log  *logger.Logger
	Path string
}

// Keystore is an interface to a simple keystore.
type Keystore interface {
	GenerateNewValidatorKey() (*bls.SecretKey, error)
	GetValidatorKey(*primitives.Worker) (*bls.SecretKey, bool)
	HasValidatorKey(*primitives.Worker) (bool, error)
	GetValidatorKeys() ([]*bls.SecretKey, error)
	Close() error
}

type Wallet struct {
	db  *badger.DB
	log *logger.Logger

	hasMaster       bool
	encryptedMaster []byte
	decryptedMaster []byte
}

var walletDBKey = []byte("master-key-encrypted")
var walletDBSalt = []byte("master-key-salt")

func NewWallet(c Config) (*Wallet, error) {
	bdb, err := badger.Open(badger.DefaultOptions(c.Path).WithLogger(nil))
	if err != nil {
		return nil, err
	}

	encryptedMaster := []byte{}
	hasMaster := false
	err = bdb.Update(func(txn *badger.Txn) error {
		i, err := txn.Get(walletDBKey)
		if err == badger.ErrKeyNotFound {
			return nil
		} else if err != nil {
			return err
		} else {
			encryptedMasterBytes, err := i.ValueCopy(nil)
			if err != nil {
				return err
			}
			encryptedMaster = encryptedMasterBytes
			hasMaster = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Wallet{
		db:              bdb,
		encryptedMaster: encryptedMaster,
		hasMaster:       hasMaster,
		log:             c.Log,
	}, nil
}

func (b *Wallet) Start() error {
	if !b.hasMaster {
		var fd int
		if terminal.IsTerminal(syscall.Stdin) {
			fd = syscall.Stdin
		} else {
			tty, err := os.Open("/dev/tty")
			if err != nil {
				return errors.Wrap(err, "error allocating terminal")
			}
			defer tty.Close()
			fd = int(tty.Fd())
		}
		fmt.Println("Creating new wallet...")
		var password []byte

		for {
			fmt.Printf("Enter a password: ")
			pass, err := terminal.ReadPassword(fd)
			if err != nil {
				return errors.Wrap(err, "error reading password")
			}
			fmt.Println()
			fmt.Printf("Re-enter the password: ")
			passVerify, err := terminal.ReadPassword(fd)
			if err != nil {
				return errors.Wrap(err, "error reading password")
			}
			fmt.Println()

			if bytes.Equal(pass, passVerify) {
				password = pass
				break
			} else {
				fmt.Println("Passwords do not match. Please try again.")
			}
		}

		// generate random salt
		var salt [8]byte
		_, err := rand.Reader.Read(salt[:])
		if err != nil {
			return errors.Wrap(err, "error reading from random")
		}
		encryptionKey := pbkdf2.Key(password, salt[:], 20000, 32, sha512.New)

		block, err := aes.NewCipher(encryptionKey)
		if err != nil {
			return errors.Wrap(err, "error creating cipher")
		}

		nonce := make([]byte, 12)
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return errors.Wrap(err, "error reading from random")
		}

		aesgcm, err := cipher.NewGCM(block)
		if err != nil {
			return errors.Wrap(err, "error creating GCM")
		}

		masterKey := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, masterKey); err != nil {
			return errors.Wrap(err, "error reading from random")
		}

		ciphertext := aesgcm.Seal(nil, nonce, masterKey, nil)

		err = b.db.Update(func(tx *badger.Txn) error {
			if err := tx.Set(walletDBKey, ciphertext[:]); err != nil {
				return err
			}

			if err := tx.Set(walletDBSalt, salt[:]); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		b.encryptedMaster = ciphertext
		b.hasMaster = true
	}

	return nil
}

func (w *Wallet) Stop() error {
	return w.db.Close()
}

func (b *Wallet) GetValidatorKeys() ([]*bls.SecretKey, error) {
	secKeys := make([]*bls.SecretKey, 0)
	err := b.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			i := iter.Item()
			val, err := i.ValueCopy(nil)
			if err != nil {
				return err
			}
			if len(val) == 32 {
				var valBytes [32]byte
				copy(valBytes[:], val)
				secretKey := bls.DeserializeSecretKey(valBytes)
				secKeys = append(secKeys, &secretKey)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return secKeys, nil
}

func (b *Wallet) GetValidatorKey(worker *primitives.Worker) (*bls.SecretKey, bool) {
	pubBytes := worker.PubKey

	var secretBytes [32]byte
	err := b.db.View(func(txn *badger.Txn) error {
		i, err := txn.Get(pubBytes[:])
		if err != nil {
			return err
		}

		_, err = i.ValueCopy(secretBytes[:])
		return err
	})
	if err != nil {
		return nil, false
	}

	secretKey := bls.DeserializeSecretKey(secretBytes)
	return &secretKey, true
}

func (b *Wallet) HasValidatorKey(worker *primitives.Worker) (result bool, err error) {
	pubBytes := worker.PubKey

	err = b.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(pubBytes[:])
		if err == badger.ErrKeyNotFound {
			result = false
		}
		if err != nil {
			return err
		}
		result = true
		return nil
	})

	return result, err
}

func (b *Wallet) GenerateNewValidatorKey() (*bls.SecretKey, error) {
	key, err := bls.RandSecretKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	keyBytes := key.Serialize()

	pub := key.DerivePublicKey()
	pubBytes := pub.Serialize()

	err = b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(pubBytes[:], keyBytes[:])
	})

	return key, err
}

func (b *Wallet) Close() error {
	return b.db.Close()
}

var _ Keystore = &Wallet{}
