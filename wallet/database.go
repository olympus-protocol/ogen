package wallet

import "go.etcd.io/bbolt"

type walletInfo struct {
	nonce     []byte
	lastNonce uint64
	address   string
}

func (w *Wallet) load() error {
	var info walletInfo
	err := w.db.Update(func(txn *bbolt.Tx) error {
		return nil
	})
	if err != nil {
		return err
	}
	w.info = info
	return nil
}
