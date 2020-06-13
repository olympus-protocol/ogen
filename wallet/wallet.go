package wallet

import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"go.etcd.io/bbolt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/mempool"
	"github.com/olympus-protocol/ogen/peers"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/logger"
)

var ErrorOpen = errors.New("please open a wallet first")

type Wallet struct {
	db            *bbolt.DB
	log           *logger.Logger
	params        *params.ChainParams
	chain         *chain.Blockchain
	mempool       *mempool.CoinsMempool
	actionMempool *mempool.ActionMempool

	txTopic      *pubsub.Topic
	depositTopic *pubsub.Topic
	exitTopic    *pubsub.Topic

	directory     string
	name          string
	open          bool
	info          walletInfo
	ctx           context.Context
	lastNonceLock sync.Mutex
}

// NewWallet creates a new wallet.
func NewWallet(ctx context.Context, log *logger.Logger, walletsDir string, params *params.ChainParams, ch *chain.Blockchain, hostnode *peers.HostNode, mempool *mempool.CoinsMempool, actionMempool *mempool.ActionMempool) (wallet *Wallet, err error) {
	var txTopic *pubsub.Topic
	var depositTopic *pubsub.Topic
	var exitTopic *pubsub.Topic
	if hostnode != nil {
		txTopic, err = hostnode.Topic("tx")
		if err != nil {
			return nil, err
		}

		depositTopic, err = hostnode.Topic("deposits")
		if err != nil {
			return nil, err
		}

		exitTopic, err = hostnode.Topic("exits")
		if err != nil {
			return nil, err
		}
	}
	wallet = &Wallet{
		log:           log,
		directory:     walletsDir,
		params:        params,
		open:          false,
		chain:         ch,
		txTopic:       txTopic,
		depositTopic:  depositTopic,
		exitTopic:     exitTopic,
		ctx:           ctx,
		mempool:       mempool,
		actionMempool: actionMempool,
	}
	return wallet, nil
}

func (w *Wallet) OpenWallet(name string) error {
	if _, err := os.Stat(path.Join(w.directory, "wallets")); os.IsNotExist(err) {
		os.Mkdir(path.Join(w.directory, "wallets"), 0700)
	}
	w.name = name
	db, err := bbolt.Open(path.Join(w.directory, "wallets", name+".db"), 0600, nil)
	if err != nil {
		//if err == bbolt.ErrInvalid {
		//	return w.hardRecover()
		//}
		return err
	}
	w.db = db
	err = w.load()
	if err == errorNotInit {
		err := w.initialize()
		if err != nil {
			return err
		}
		err = w.load()
		if err != nil {
			return err
		}
		w.open = true
		return nil
	}
	if err == errorNoInfo {
		err := w.recover()
		//if err != nil {
		//	return w.hardRecover()
		//}
		return err
	}
	//if err != nil {
	//	return w.hardRecover()
	//}
	w.open = true
	return nil
}

func (w *Wallet) CloseWallet() error {
	w.open = false
	w.info = walletInfo{}
	w.name = ""
	return w.db.Close()
}

func (w *Wallet) HasWallet(name string) bool {
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

func (w *Wallet) GetAvailableWallets() (map[string]string, error) {
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

func (w *Wallet) GetAccount() (string, error) {
	if !w.open {
		return "", errorNotOpen
	}
	return bech32.Encode(w.params.AddrPrefix.Public, w.info.account[:]), nil
}

func (w *Wallet) GetAccountRaw() ([20]byte, error) {
	if !w.open {
		return [20]byte{}, errorNotOpen
	}
	if len(w.info.account) != 20 {
		return [20]byte{}, errors.New("expected address to be 20 bytes")
	}
	return w.info.account, nil
}
