package walletdb

import (
	"github.com/grupokindynos/ogen/utils/serializer"
	"io"
	"sync"
)

var walletCredentialsBucketKey = []byte("wallet-credentials")

type WalletCredentials struct {
	lock     sync.RWMutex
	Accounts map[int32]Account
	Mnemonic string
}

type Account struct {
	Number            int32
	Path              string
	ExtendedPublicKey string
}

func (acc *Account) serialize(w io.Writer) error {
	err := serializer.WriteElement(w, acc.Number)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, acc.Path)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, acc.ExtendedPublicKey)
	if err != nil {
		return err
	}
	return nil
}

func (acc *Account) deserialize(r io.Reader) error {
	err := serializer.ReadElement(r, &acc.Number)
	if err != nil {
		return err
	}
	acc.Path, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	acc.ExtendedPublicKey, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	return nil
}

func (cred *WalletCredentials) Serialize(w io.Writer) error {
	err := serializer.WriteVarInt(w, uint64(len(cred.Accounts)))
	if err != nil {
		return err
	}
	for _, acc := range cred.Accounts {
		err = acc.serialize(w)
		if err != nil {
			return err
		}
	}
	err = serializer.WriteVarString(w, cred.Mnemonic)
	if err != nil {
		return err
	}
	return nil
}

func (cred *WalletCredentials) Deserialize(r io.Reader) error {
	count, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	cred.Accounts = make(map[int32]Account, count)
	for i := uint64(0); i < count; i++ {
		var acc Account
		err := acc.deserialize(r)
		if err != nil {
			return err
		}
		cred.AddAccount(acc)
	}
	cred.Mnemonic, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	return nil
}

func (cred *WalletCredentials) AddAccount(account Account) {
	cred.lock.Lock()
	cred.Accounts[account.Number] = account
	cred.lock.Unlock()
	return
}
