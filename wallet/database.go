package wallet

import (
	"encoding/binary"

	"github.com/dgraph-io/badger"
)

type walletInfo struct {
	encryptedMaster []byte
	salt            []byte
	nonce           []byte
	lastNonce       uint64
	address         string
}

func (w *Wallet) loadFromDisk() error {
	var info walletInfo
	hasMaster := false

	err := w.db.Update(func(txn *badger.Txn) error {
		// we try to read or error first
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

		// extract the values
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

		// if we are still good, update the info
		info.lastNonce = binary.BigEndian.Uint64(txNonceBytes)
		info.salt = saltBytes
		info.encryptedMaster = encryptedMasterBytes
		info.nonce = nonceBytes
		info.address = string(addressBytes)

		hasMaster = true
		return nil
	})
	if err != nil {
		return err
	}

	w.hasMaster = hasMaster
	w.info = info

	return nil
}
