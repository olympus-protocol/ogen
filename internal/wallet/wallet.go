package wallet

import (
	"context"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/bip39"
	"github.com/olympus-protocol/ogen/pkg/hdwallet"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.etcd.io/bbolt"

	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/pkg/bls"

	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
)

const defaultWalletPath = "m/12381/1997/0/0"

// Wallet is the interface for wallet
type Wallet interface {
	NewWallet(name string, mnemonic string, password string) error
	OpenWallet(name string, password string) error
	CloseWallet() error
	HasWallet(name string) bool
	GetAvailableWallets() (map[string]string, error)
	GetAccount() (string, error)
	GetSecret() (*bls.SecretKey, error)
	GetMnemonic() (string, error)
	GetPublic() (*bls.PublicKey, error)
	GetAccountRaw() ([20]byte, error)
	GetBalance() (uint64, uint64, error)
	StartValidatorBulk(k []*bls.SecretKey) (bool, error)
	ExitValidatorBulk(k []*bls.PublicKey) (bool, error)
	StartValidator(validatorPrivBytes *bls.SecretKey) (bool, error)
	ExitValidator(validatorPubKey *bls.PublicKey) (bool, error)
	SendToAddress(to string, amount uint64) (*chainhash.Hash, error)
}

var _ Wallet = &wallet{}

// wallet is the structure of the wallet manager.
type wallet struct {
	// Wallet manager properties
	netParams *params.ChainParams
	log       logger.Logger
	chain     chain.Blockchain
	host      hostnode.HostNode

	coinsmempool   mempool.CoinsMempool
	actionsmempool mempool.ActionMempool

	directory string
	ctx       context.Context

	// Open wallet information
	db         *bbolt.DB
	name       string
	open       bool
	priv       *bls.SecretKey
	pub        *bls.PublicKey
	accountRaw [20]byte
	account    string
}

// NewWallet creates a new wallet.
func NewWallet(ch chain.Blockchain, hostnode hostnode.HostNode, mempool mempool.CoinsMempool, actionMempool mempool.ActionMempool) (Wallet, error) {
	netParams := config.GlobalParams.NetParams
	ctx := config.GlobalParams.Context
	log := config.GlobalParams.Logger

	wall := &wallet{
		log:            log,
		directory:      config.GlobalFlags.DataPath,
		netParams:      netParams,
		open:           false,
		chain:          ch,
		host:           hostnode,
		coinsmempool:   mempool,
		ctx:            ctx,
		actionsmempool: actionMempool,
	}
	return wall, nil
}

// NewWallet creates a new wallet database.
func (w *wallet) NewWallet(name string, mnemonic string, password string) error {
	if w.open {
		w.CloseWallet()
	}
	passhash := chainhash.HashH([]byte(password))

	var mnemonicPhrase string
	var err error
	if mnemonic == "" {
		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			return err
		}
		mnemonicPhrase, err = bip39.NewMnemonic(entropy)
	} else {
		if !bip39.IsMnemonicValid(mnemonic) {
			return bip39.ErrInvalidMnemonic
		}
		mnemonicPhrase = mnemonic
	}

	if _, err := os.Stat(path.Join(w.directory, "wallets")); os.IsNotExist(err) {
		_ = os.Mkdir(path.Join(w.directory, "wallets"), 0700)
	}
	db, err := bbolt.Open(path.Join(w.directory, "wallets", name+".db"), 0600, nil)
	if err != nil {
		return err
	}

	seed := bip39.NewSeed(mnemonicPhrase, password)

	secret, err := hdwallet.CreateHDWallet(seed, defaultWalletPath)
	if err != nil {
		return err
	}

	w.db = db
	w.name = name
	w.priv = secret
	w.open = true
	w.pub = secret.PublicKey()
	w.account = w.pub.ToAccount()
	w.accountRaw, err = w.pub.Hash()
	if err != nil {
		return err
	}
	return w.initialize(passhash, mnemonicPhrase)
}

// OpenWallet opens an already created wallet database.
func (w *wallet) OpenWallet(name string, password string) error {
	if w.open {
		w.CloseWallet()
	}
	db, err := bbolt.Open(path.Join(w.directory, "wallets", name+".db"), 0600, nil)
	if err != nil {
		return err
	}
	w.db = db
	w.name = name
	secret, err := w.getSecret(password)
	if err != nil {
		_ = db.Close()
		w.db = nil
		w.name = ""
		return err
	}
	w.priv = secret
	w.pub = secret.PublicKey()
	w.account = w.pub.ToAccount()
	w.accountRaw, err = w.pub.Hash()
	if err != nil {
		return err
	}
	w.open = true
	return nil
}

// CloseWallet closes the current opened wallet.
func (w *wallet) CloseWallet() error {
	w.open = false
	w.name = ""
	w.priv = nil
	w.pub = nil
	w.account = ""
	w.accountRaw = [20]byte{}
	return w.db.Close()
}

// HasWallet checks if the name matches to an existing wallet database.
func (w *wallet) HasWallet(name string) bool {
	list, err := w.GetAvailableWallets()
	if err != nil {
		return false
	}
	if len(list) == 0 {
		return false
	}
	_, ok := list[name]
	return ok
}

// GetAvailableWallets returns a map of available wallets.
func (w *wallet) GetAvailableWallets() (map[string]string, error) {
	files := map[string]string{}
	err := filepath.Walk(path.Join(w.directory, "wallets/"), func(path string, info os.FileInfo, err error) error {
		if info != nil {
			if !info.IsDir() {
				if filepath.Ext(path) == ".db" {
					name := strings.Split(info.Name(), ".db")
					files[name[0]] = path + "/" + info.Name()
				}
			}
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// GetAccount returns the current wallet account on bech32 format.
func (w *wallet) GetAccount() (string, error) {
	if !w.open {
		return "", errorNotOpen
	}
	return w.account, nil
}

// GetSecret returns the secret key of the current wallet.
func (w *wallet) GetSecret() (*bls.SecretKey, error) {
	if !w.open {
		return nil, errorNotOpen
	}
	return w.priv, nil
}

// GetSecret returns the secret key of the current wallet.
func (w *wallet) GetMnemonic() (string, error) {
	if !w.open {
		return "", errorNotOpen
	}
	return w.getMnemonic()
}

// GetPublic returns the public key of the current wallet.
func (w *wallet) GetPublic() (*bls.PublicKey, error) {
	if !w.open {
		return nil, errorNotOpen
	}
	return w.pub, nil
}

// GetAccountRaw returns the current wallet account on a bytes slice.
func (w *wallet) GetAccountRaw() ([20]byte, error) {
	if !w.open {
		return [20]byte{}, errorNotOpen
	}
	return w.accountRaw, nil
}
