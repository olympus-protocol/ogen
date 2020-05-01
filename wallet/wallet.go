package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/btcsuite/btcutil/bech32"
	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
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
	db     *badger.DB
	log    *logger.Logger
	params *params.ChainParams
	chain  *chain.Blockchain

	hasMaster       bool
	encryptedMaster []byte
	salt            []byte
	nonce           []byte

	masterPriv atomic.Value
	masterLock uint32

	lastNonce     uint64
	lastNonceLock sync.Mutex

	walletAddress string
}

var walletDBKey = []byte("encryption-key-ciphertext")
var walletDBSalt = []byte("encryption-key-salt")
var walletDBNonce = []byte("encryption-key-nonce")
var walletDBLastTxNonce = []byte("last-tx-nonce")
var walletDBAddress = []byte("wallet-address")

func NewWallet(c Config, params params.ChainParams, ch *chain.Blockchain) (*Wallet, error) {
	bdb, err := badger.Open(badger.DefaultOptions(c.Path).WithLogger(nil))
	if err != nil {
		return nil, err
	}

	var encryptedMaster []byte
	var salt []byte
	var nonce []byte
	var lastNonce uint64
	var address string

	hasMaster := false
	err = bdb.Update(func(txn *badger.Txn) error {
		masterItem, err := txn.Get(walletDBKey)
		if err == badger.ErrKeyNotFound {
			return nil
		} else if err != nil {
			return err
		}
		saltItem, err := txn.Get(walletDBSalt)
		if err == badger.ErrKeyNotFound {
			return nil
		} else if err != nil {
			return err
		}
		nonceItem, err := txn.Get(walletDBNonce)
		if err == badger.ErrKeyNotFound {
			return nil
		} else if err != nil {
			return err
		}
		txNonce, err := txn.Get(walletDBLastTxNonce)
		if err == badger.ErrKeyNotFound {
			return nil
		} else if err != nil {
			return err
		}
		addressItem, err := txn.Get(walletDBAddress)
		if err == badger.ErrKeyNotFound {
			return nil
		} else if err != nil {
			return err
		}
		addressBytes, err := addressItem.ValueCopy(nil)
		if err != nil {
			return err
		}
		encryptedMasterBytes, err := masterItem.ValueCopy(nil)
		if err != nil {
			return err
		}
		saltBytes, err := saltItem.ValueCopy(nil)
		if err != nil {
			return err
		}
		nonceBytes, err := nonceItem.ValueCopy(nil)
		if err != nil {
			return err
		}
		txNonceBytes, err := txNonce.ValueCopy(nil)
		if err != nil {
			return err
		}
		lastNonce = binary.BigEndian.Uint64(txNonceBytes)
		salt = saltBytes
		encryptedMaster = encryptedMasterBytes
		nonce = nonceBytes
		hasMaster = true
		address = string(addressBytes)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Wallet{
		db:              bdb,
		encryptedMaster: encryptedMaster,
		salt:            salt,
		nonce:           nonce,
		lastNonce:       lastNonce,
		hasMaster:       hasMaster,
		log:             c.Log,
		params:          &params,
		walletAddress:   address,
		chain:           ch,
	}, nil
}

func (b *Wallet) unlock(authentication []byte) error {
	if !atomic.CompareAndSwapUint32(&b.masterLock, 0, 1) {
		return nil
	}

	encryptionKey := pbkdf2.Key(authentication, b.salt, 20000, 32, sha512.New)

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

	masterSeed, err := aesgcm.Open(nil, b.nonce, b.encryptedMaster, nil)
	if err != nil {
		return errors.Wrap(err, "could not decrypt master key")
	}

	var secretKeyBytes [32]byte
	copy(secretKeyBytes[:], masterSeed)

	secKey := bls.DeriveSecretKey(secretKeyBytes)

	b.masterPriv.Store(secKey)

	go func() {
		<-time.After(time.Minute * 2)
		b.masterPriv.Store(nil)
		atomic.StoreUint32(&b.masterLock, 0)
	}()

	return nil
}

func (b *Wallet) unlockIfNeeded(authentication []byte) (*bls.SecretKey, error) {
	privVal := b.masterPriv.Load()
	if privVal == nil {
		if authentication == nil {
			return nil, fmt.Errorf("wallet locked, need authentication")
		}

		if err := b.unlock(authentication); err != nil {
			return nil, err
		}

		privVal = b.masterPriv.Load()
	}

	return privVal.(*bls.SecretKey), nil
}

func (b *Wallet) GetAddress() string {
	return b.walletAddress
}

func (b *Wallet) SendToAddress(authentication []byte, to string, amount uint64) (*chainhash.Hash, error) {
	priv, err := b.unlockIfNeeded(authentication)
	if err != nil {
		return nil, err
	}

	_, data, err := bech32.Decode(to)
	if err != nil {
		return nil, err
	}

	if len(data) != 20 {
		return nil, fmt.Errorf("invalid address")
	}

	var toPkh [20]byte

	copy(toPkh[:], data)

	pub := priv.DerivePublicKey()

	b.lastNonceLock.Lock()
	b.lastNonce++
	nonce := b.lastNonce
	b.lastNonceLock.Unlock()

	payload := &primitives.CoinPayload{
		To:            toPkh,
		FromPublicKey: *pub,
		Amount:        amount,
		Nonce:         nonce,
		Fee:           100,
	}

	sigMsg := payload.SignatureMessage()
	sig, err := bls.Sign(priv, sigMsg[:])
	if err != nil {
		return nil, err
	}

	payload.Signature = *sig

	tx := &primitives.Tx{
		TxType:    primitives.TxCoins,
		TxVersion: 0,
		Payload:   payload,
	}

	if err := b.chain.SubmitCoinTransaction(payload); err != nil {
		return nil, err
	}

	txHash := tx.Hash()

	return &txHash, nil
}

func (b *Wallet) Start() error {
	var password []byte

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

		var privateKeyBytes [32]byte
		if _, err := io.ReadFull(rand.Reader, privateKeyBytes[:]); err != nil {
			return errors.Wrap(err, "error reading from random")
		}

		privateKey := bls.DeriveSecretKey(privateKeyBytes)
		address, err := privateKey.DerivePublicKey().ToBech32(b.params.AddressPrefixes)
		if err != nil {
			return errors.Wrap(err, "could not get public key from private key")
		}

		ciphertext := aesgcm.Seal(nil, nonce, privateKeyBytes[:], nil)

		err = b.db.Update(func(tx *badger.Txn) error {
			if err := tx.Set(walletDBKey, ciphertext[:]); err != nil {
				return err
			}

			if err := tx.Set(walletDBAddress, []byte(address)); err != nil {
				return err
			}

			if err := tx.Set(walletDBSalt, salt[:]); err != nil {
				return err
			}

			if err := tx.Set(walletDBLastTxNonce, []byte{0, 0, 0, 0, 0, 0, 0, 0}); err != nil {
				return err
			}

			return tx.Set(walletDBNonce, nonce[:])
		})
		if err != nil {
			return err
		}

		b.encryptedMaster = ciphertext
		b.salt = salt[:]
		b.nonce = nonce
		b.lastNonce = 0
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
