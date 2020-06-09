package wallet

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"path"

	"github.com/olympus-protocol/ogen/bls"
	"go.etcd.io/bbolt"
)

var privKeyMagicBytes = []byte{0x53, 0xB3, 0x31, 0x0F}

var errorNotInit = errors.New("the wallet is not initialized")
var errorNoInfo = errors.New("wallet corruption, some elements are not found on the wallet")
var errorNotOpen = errors.New("there is no wallet open, please open one first")

type walletInfo struct {
	nonce     []byte
	lastNonce uint64
	account   string
}

var walletInfoBucketKey = []byte("wallet_info")
var walletAccountDbKey = []byte("address")
var walletNonceDbKey = []byte("nonce")
var walletLastNonceDbKey = []byte("last_nonce")
var walletKeyBucket = []byte("keys")
var walletPrivKeyDbKey = []byte("priv_key")

func (w *Wallet) load() error {
	var loadInfo walletInfo
	err := w.db.View(func(txn *bbolt.Tx) error {
		info := txn.Bucket(walletInfoBucketKey)
		if info == nil {
			return errorNotInit
		}
		address := string(info.Get(walletAccountDbKey))
		if address == "" {
			return errorNoInfo
		}
		nonce := info.Get(walletNonceDbKey)
		if nonce == nil {
			return errorNoInfo
		}
		lastNonce := binary.LittleEndian.Uint64(info.Get(walletLastNonceDbKey))
		if lastNonce < 0 {
			return errorNoInfo
		}
		loadInfo = walletInfo{
			account:   address,
			nonce:     nonce,
			lastNonce: lastNonce,
		}
		return nil
	})
	if err != nil {
		return err
	}
	w.info = loadInfo
	return nil
}

func (w *Wallet) GetPublicKey() (string, error) {
	if !w.open {
		return "", errorNotOpen
	}
	return w.info.account, nil
}

func (w *Wallet) GetSecret() (*bls.SecretKey, error) {
	var secret *bls.SecretKey
	err := w.db.View(func(tx *bbolt.Tx) error {
		keyBucket := tx.Bucket(walletKeyBucket)
		if keyBucket == nil {
			return errors.New("no key bucket available")
		}
		privKeyBytesSet := keyBucket.Get(walletPrivKeyDbKey)
		if privKeyBytesSet == nil {
			return errors.New("no private key value available")
		}
		privKeyBytesSlice := bytes.Split(privKeyBytesSet, privKeyMagicBytes)
		privKeyBytes := privKeyBytesSlice[1]
		secret = bls.DeriveSecretKey(privKeyBytes)
		if secret == nil {
			return errors.New("no private key found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (w *Wallet) recover() error {
	db, err := bbolt.Open(path.Join(w.directory, "wallets", w.name+".db"), 0600, nil)
	if err != nil {
		return err
	}
	db.Update(func(tx *bbolt.Tx) error {
		// Get the private Key to ensure it is there
		keyBucket := tx.Bucket(walletKeyBucket)
		if keyBucket == nil {
			return errors.New("no key bucket available")
		}
		privKeyBytesSet := keyBucket.Get(walletPrivKeyDbKey)
		if privKeyBytesSet == nil {
			return errors.New("no private key value available")
		}
		privKeyBytesSlice := bytes.Split(privKeyBytesSet, privKeyMagicBytes)
		privKeyBytes := privKeyBytesSlice[1]
		blsPrivKey := bls.DeriveSecretKey(privKeyBytes)
		if blsPrivKey == nil {
			return errors.New("no private key found")
		}
		// If the private key is available, just reinitialize the info bucket
		infoBucket, err := tx.CreateBucketIfNotExists(walletInfoBucketKey)
		if err != nil {
			return err
		}
		nonce := make([]byte, 12)
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return errors.New("error reading from random" + err.Error())
		}
		lastNonce := []byte{0, 0, 0, 0, 0, 0, 0, 0}
		account, err := blsPrivKey.PublicKey().ToAddress(w.params.AddrPrefix.Public)
		if err != nil {
			return err
		}
		err = infoBucket.Put(walletNonceDbKey, nonce)
		if err != nil {
			return err
		}
		err = infoBucket.Put(walletLastNonceDbKey, lastNonce)
		if err != nil {
			return err
		}
		err = infoBucket.Put(walletAccountDbKey, []byte(account))
		if err != nil {
			return err
		}
		return nil
	})
	return nil
}

func (w *Wallet) initialize() error {
	privateKey := bls.RandKey()
	w.db.Update(func(tx *bbolt.Tx) error {
		keyBucket, err := tx.CreateBucketIfNotExists(walletKeyBucket)
		if err != nil {
			return err
		}
		var encapsulatedPrivKey []byte
		encapsulatedPrivKey = append(encapsulatedPrivKey, privKeyMagicBytes...)
		encapsulatedPrivKey = append(encapsulatedPrivKey, privateKey.Marshal()...)
		encapsulatedPrivKey = append(encapsulatedPrivKey, privKeyMagicBytes...)
		err = keyBucket.Put(walletPrivKeyDbKey, encapsulatedPrivKey)
		if err != nil {
			return err
		}
		infoBucket, err := tx.CreateBucketIfNotExists(walletInfoBucketKey)
		if err != nil {
			return err
		}
		nonce := make([]byte, 12)
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return errors.New("error reading from random" + err.Error())
		}
		lastNonce := []byte{0, 0, 0, 0, 0, 0, 0, 0}
		account, err := privateKey.PublicKey().ToAddress(w.params.AddrPrefix.Public)
		if err != nil {
			return err
		}
		err = infoBucket.Put(walletNonceDbKey, nonce)
		if err != nil {
			return err
		}
		err = infoBucket.Put(walletLastNonceDbKey, lastNonce)
		if err != nil {
			return err
		}
		err = infoBucket.Put(walletAccountDbKey, []byte(account))
		if err != nil {
			return err
		}
		return nil
	})
	return nil
}

func (w *Wallet) hardRecover() error {
	// file, err := os.Open(path.Join(w.directory, "wallets", w.name + ".db"))
	// if err != nil {
	// 	return err
	// }
	// stats, err := file.Stat()
	// if err != nil {
	// 	return err
	// }
	// fileSize := stats.Size()
	// fileBytes := make([]byte, fileSize)
	// _, _ = file.ReadAt(fileBytes, 0)
	// // Split the entire file and look for the privKeyMagicBytes
	// // If this is successfull the privKey should be on position [1] of the byte array
	// fmt.Println(file)
	// splitBytesSet := bytes.Split(fileBytes, privKeyMagicBytes)
	// fmt.Println(splitBytesSet)
	// privkey := splitBytesSet[1]
	// blsPrivKey := bls.DeriveSecretKey(privkey)
	// if blsPrivKey == nil {
	// 	return errors.New("unable to recover private key")
	// }
	// // We recovered the private key, we should remove the file and initialize the database
	// _ = os.Remove(path.Join(w.directory, "wallets", w.name))
	// w.db, err = bbolt.Open(path.Join(w.directory, "wallets", w.name+".db"), 0600, nil)
	// if err != nil {
	// 	return err
	// }
	// w.db.Update(func(tx *bbolt.Tx) error {
	// 	keyBucket, err := tx.CreateBucketIfNotExists(walletKeyBucket)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	var encapsulatedPrivKey []byte
	// 	copy(encapsulatedPrivKey, privKeyMagicBytes)
	// 	copy(encapsulatedPrivKey, blsPrivKey.Marshal())
	// 	copy(encapsulatedPrivKey, privKeyMagicBytes)
	// 	err = keyBucket.Put(walletPrivKeyDbKey, encapsulatedPrivKey)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	infoBucket, err := tx.CreateBucketIfNotExists(walletInfoBucketKey)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	nonce := make([]byte, 12)
	// 	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
	// 		return errors.Wrap(err, "error reading from random")
	// 	}
	// 	lastNonce := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	// 	account, err := blsPrivKey.PublicKey().ToAddress(w.params.AddrPrefix.Public)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = infoBucket.Put(walletNonceDbKey, nonce)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = infoBucket.Put(walletLastNonceDbKey, lastNonce)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = infoBucket.Put(walletAccountDbKey, []byte(account))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return nil
	// })
	// return nil
	return nil
}
