package wallet

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/blsaes"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"go.etcd.io/bbolt"
)

var privKeyMagicBytes = []byte{0x53, 0xB3, 0x31, 0x0F}

var errorNotInit = errors.New("the wallet is not initialized")
var errorNoInfo = errors.New("wallet corruption, some elements are not found on the wallet")
var errorNotOpen = errors.New("there is no wallet open, please open one first")

var walletKeyBucket = []byte("keys")
var walletPrivKeyDbKey = []byte("private")

var walletInfoBucket = []byte("info")
var walletPassHashDbKey = []byte("passhash")
var walletSaltDbKey = []byte("salt")
var walletNonceDbKey = []byte("nonce")

func (w *Wallet) initialize(cipher []byte, salt []byte, nonce []byte, passhash chainhash.Hash) error {
	return w.db.Update(func(tx *bbolt.Tx) error {
		keybkt, err := tx.CreateBucket(walletKeyBucket)
		if err != nil {
			return err
		}
		var encKeyCipher []byte
		encKeyCipher = append(encKeyCipher, privKeyMagicBytes...)
		encKeyCipher = append(encKeyCipher, cipher...)
		encKeyCipher = append(encKeyCipher, privKeyMagicBytes...)
		err = keybkt.Put(walletPrivKeyDbKey, encKeyCipher)
		if err != nil {
			return err
		}
		infobkt, err := tx.CreateBucket(walletInfoBucket)
		if err != nil {
			return err
		}
		err = infobkt.Put(walletSaltDbKey, salt)
		if err != nil {
			return err
		}
		err = infobkt.Put(walletNonceDbKey, nonce)
		if err != nil {
			return err
		}
		err = infobkt.Put(walletPassHashDbKey, passhash[:])
		if err != nil {
			return err
		}
		return nil
	})
}

func (w *Wallet) getSecret(password string) (key *bls.SecretKey, err error) {
	err = w.db.View(func(tx *bbolt.Tx) error {
		infobkt := tx.Bucket(walletInfoBucket)
		currPassHash := chainhash.HashB([]byte(password))
		passhash := infobkt.Get(walletPassHashDbKey)
		equal := reflect.DeepEqual(currPassHash, passhash)
		if !equal {
			return errors.New("password don't match")
		}
		salt := infobkt.Get(walletSaltDbKey)
		nonce := infobkt.Get(walletNonceDbKey)
		keybkt := tx.Bucket(walletKeyBucket)
		cipherBytesSet := keybkt.Get(walletPrivKeyDbKey)
		if cipherBytesSet == nil {
			return errors.New("no private key value available")
		}
		cipherBytesSlice := bytes.Split(cipherBytesSet, privKeyMagicBytes)
		cipherBytes := cipherBytesSlice[1]
		key, err = blsaes.Decrypt(cipherBytes, nonce, []byte(password), salt)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}
