package wallet

import (
	"errors"
	"github.com/olympus-protocol/ogen/pkg/bip39"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/hdwallet"
	"reflect"

	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"go.etcd.io/bbolt"
)

var (
	errorNotOpen = errors.New("there is no wallet open, please open one first")

	walletMnemonicBucket = []byte("mnemonic")
	walletInfoBucket     = []byte("info")

	walletPassHashDbKey = []byte("passhash")
	walletMnemonicKey   = []byte("mnemonic")
)

func (w *wallet) initialize(passhash chainhash.Hash, mnemonic string) error {

	return w.db.Update(func(tx *bbolt.Tx) error {
		keybkt, err := tx.CreateBucket(walletMnemonicBucket)
		if err != nil {
			return err
		}

		infobkt, err := tx.CreateBucket(walletInfoBucket)
		if err != nil {
			return err
		}

		err = keybkt.Put(walletMnemonicKey, []byte(mnemonic))
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

		mnemonic := tx.Bucket(walletMnemonicBucket).Get(walletMnemonicKey)

		seed := bip39.NewSeed(string(mnemonic), password)

		key, err = hdwallet.CreateHDWallet(seed, defaultWalletPath)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (w *wallet) getMnemonic() (mnemonic string, err error) {
	err = w.db.View(func(tx *bbolt.Tx) error {

		mnemonicBytes := tx.Bucket(walletMnemonicBucket).Get(walletMnemonicKey)

		mnemonic = string(mnemonicBytes)

		return nil
	})
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}
