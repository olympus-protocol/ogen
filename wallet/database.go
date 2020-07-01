package wallet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"path"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"go.etcd.io/bbolt"
)

var privKeyMagicBytes = []byte{0x53, 0xB3, 0x31, 0x0F}

var errorNotInit = errors.New("the wallet is not initialized")
var errorNoInfo = errors.New("wallet corruption, some elements are not found on the wallet")
var errorNotOpen = errors.New("there is no wallet open, please open one first")

type walletInfo struct {
	nonce     uint64
	lastNonce uint64
	account   [20]byte
}

var walletInfoBucketKey = []byte("wallet_info")
var walletAccountDbKey = []byte("address")
var walletNonceDbKey = []byte("nonce")
var walletKeyBucket = []byte("keys")
var walletPrivKeyDbKey = []byte("priv_key")

func (w *Wallet) load() error {
	var loadInfo walletInfo
	err := w.db.View(func(txn *bbolt.Tx) error {
		info := txn.Bucket(walletInfoBucketKey)
		if info == nil {
			return errorNotInit
		}
		var account [20]byte
		stAcc := info.Get(walletAccountDbKey)
		if stAcc == nil {
			return errorNoInfo
		}
		if len(stAcc) < 20 {
			return errorNoInfo
		}
		copy(account[:], stAcc)
		nonce := info.Get(walletNonceDbKey)
		if nonce == nil {
			return errorNoInfo
		}
		lastNonce := binary.LittleEndian.Uint64(nonce)
		if lastNonce < 0 {
			return errorNoInfo
		}
		loadInfo = walletInfo{
			account: account,
			nonce:   lastNonce,
		}
		return nil
	})
	if err != nil {
		return err
	}
	w.info = loadInfo
	return nil
}

// func (w *Wallet) SetNonce(nonce uint64) error {
// 	err = db.Update(func(tx *bbolt.Tx) error {

// 		return nil
// 	}
// 	return err
// }

func (w *Wallet) GetSecret() (s *bls.SecretKey, err error) {
	if !w.open {
		return nil, errorNotOpen
	}
	err = w.db.View(func(tx *bbolt.Tx) error {
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
		s, err = bls.SecretKeyFromBytes(privKeyBytes)
		if err != nil {
			return errors.New("no private key found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s, nil
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
		blsPrivKey, err := bls.SecretKeyFromBytes(privKeyBytes)
		if blsPrivKey == nil {
			return errors.New("no private key found")
		}
		// If the private key is available, just reinitialize the info bucket
		infoBucket, err := tx.CreateBucketIfNotExists(walletInfoBucketKey)
		if err != nil {
			return err
		}
		account, err := blsPrivKey.PublicKey().ToAddress(w.params.AddrPrefix.Public)
		if err != nil {
			return err
		}
		nonce := make([]byte, 8)
		binary.LittleEndian.PutUint64(nonce, 0)
		err = infoBucket.Put(walletNonceDbKey, nonce)
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

func (w *Wallet) initialize(prv *bls.SecretKey) error {
	w.db.Update(func(tx *bbolt.Tx) error {
		keyBucket, err := tx.CreateBucketIfNotExists(walletKeyBucket)
		if err != nil {
			return err
		}
		var encapsulatedPrivKey []byte
		encapsulatedPrivKey = append(encapsulatedPrivKey, privKeyMagicBytes...)
		encapsulatedPrivKey = append(encapsulatedPrivKey, prv.Marshal()...)
		encapsulatedPrivKey = append(encapsulatedPrivKey, privKeyMagicBytes...)
		err = keyBucket.Put(walletPrivKeyDbKey, encapsulatedPrivKey)
		if err != nil {
			return err
		}
		infoBucket, err := tx.CreateBucketIfNotExists(walletInfoBucketKey)
		if err != nil {
			return err
		}
		pubKey := prv.PublicKey()
		pubKeyBytes := pubKey.Marshal()
		var account [20]byte
		h := chainhash.HashH(pubKeyBytes[:])
		copy(account[:], h[:20])
		err = infoBucket.Put(walletAccountDbKey, account[:])
		if err != nil {
			return err
		}
		nonce := make([]byte, 8)
		binary.LittleEndian.PutUint64(nonce, 0)
		err = infoBucket.Put(walletNonceDbKey, nonce)
		if err != nil {
			return err
		}
		return nil
	})
	return nil
}
