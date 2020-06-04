package wallet

import (
	"errors"
	"path"
	"sync"

	"go.etcd.io/bbolt"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/logger"
)

var ErrorOpen = errors.New("please open a wallet first")

type Config struct {
	Log  *logger.Logger
	Path string
}

type Wallet struct {
	db            *bbolt.DB
	config        Config
	log           *logger.Logger
	params        *params.ChainParams
	open          bool
	info          walletInfo
	lastNonceLock sync.Mutex
}

// NewWallet creates a new wallet.
func NewWallet(conf Config, params params.ChainParams) (*Wallet, error) {

	w := &Wallet{
		log:    conf.Log,
		config: conf,
		params: &params,
		open:   false,
	}

	if err := w.load(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Wallet) OpenWallet(name string) error {
	db, err := bbolt.Open(path.Join(w.config.Path, name+".db"), 0600, nil)
	if err != nil {
		return err
	}
	w.db = db
	w.open = true
	return w.initializeWallet()
}

func (w *Wallet) CloseWallet() error {
	w.open = false
	return w.db.Close()
}
