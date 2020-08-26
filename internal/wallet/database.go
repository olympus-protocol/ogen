package wallet

import (
	"bytes"
	"errors"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"reflect"

	"github.com/olympus-protocol/ogen/pkg/aesbls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"go.etcd.io/bbolt"
)

var (
	errorNotOpen = errors.New("there is no wallet open, please open one first")

	privKeyMagicBytes = []byte{0x53, 0xB3, 0x31, 0x0F}

	walletKeyBucket  = []byte("keys")
	walletInfoBucket = []byte("info")

	walletPrivKeyDbKey  = []byte("private")
	walletPassHashDbKey = []byte("passhash")
	walletSaltDbKey     = []byte("salt")
	walletNonceDbKey    = []byte("nonce")
)

func (w *wallet) initialize(cipher []byte, salt [8]byte, nonce [12]byte, passhash chainhash.Hash) error {
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

		err = infobkt.Put(walletSaltDbKey, salt[:])
		if err != nil {
			return err
		}

		err = infobkt.Put(walletNonceDbKey, nonce[:])
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

func (w *wallet) getSecret(password string) (key *bls.SecretKey, err error) {
	err = w.db.View(func(tx *bbolt.Tx) error {

		infobkt := tx.Bucket(walletInfoBucket)

		currPassHash := chainhash.HashB([]byte(password))

		passhash := infobkt.Get(walletPassHashDbKey)

		equal := reflect.DeepEqual(currPassHash, passhash)
		if !equal {
			return errors.New("password don't match")
		}

		var salt [8]byte
		var nonce [12]byte

		saltB := infobkt.Get(walletSaltDbKey)
		nonceB := infobkt.Get(walletNonceDbKey)

		copy(salt[:], saltB)
		copy(nonce[:], nonceB)

		keybkt := tx.Bucket(walletKeyBucket)

		cipherBytesSet := keybkt.Get(walletPrivKeyDbKey)

		if cipherBytesSet == nil {
			return errors.New("no private key value available")
		}

		cipherBytesSlice := bytes.Split(cipherBytesSet, privKeyMagicBytes)
		cipherBytes := cipherBytesSlice[1]

		key, err = aesbls.Decrypt(nonce, salt, cipherBytes, []byte(password))
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
