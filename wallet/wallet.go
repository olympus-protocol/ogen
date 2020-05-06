package wallet

import (
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/utils/hdwallets"
)

var PolisNetPrefix = &hdwallets.NetPrefix{
	ExtPub:  []byte{0x1f, 0x74, 0x90, 0xf0},
	ExtPriv: []byte{0x11, 0x24, 0xd9, 0x70},
}

type Config struct {
	Log  *logger.Logger
	Path string
}

type Wallet struct {
	db     *badger.DB
	log    *logger.Logger
	params *params.ChainParams

	chain   *chain.Blockchain
	peerman *peers.PeerMan

	hasMaster bool

	info          walletInfo
	lastNonceLock sync.Mutex
}

var walletDBKey = []byte("encryption-key-ciphertext")
var walletDBSalt = []byte("encryption-key-salt")
var walletDBNonce = []byte("encryption-key-nonce")
var walletDBLastTxNonce = []byte("last-tx-nonce")
var walletDBAddress = []byte("wallet-address")

func NewWallet(c Config, params params.ChainParams, ch *chain.Blockchain, peerman *peers.PeerMan) (*Wallet, error) {
	bdb, err := badger.Open(badger.DefaultOptions(c.Path).WithLogger(nil))
	if err != nil {
		return nil, err
	}

	w := &Wallet{
		db:        bdb,
		hasMaster: false,
		log:       c.Log,
		params:    &params,
		chain:     ch,
		peerman:   peerman,
	}

	if err := w.loadFromDisk(); err != nil {
		return nil, err
	}

	return w, nil
}

func (b *Wallet) Start() error {
	if !b.hasMaster {
		return b.initializeWallet()
	}

	return nil
}

func (w *Wallet) Stop() error {
	return w.db.Close()
}

func (b *Wallet) Close() error {
	return b.db.Close()
}

var _ Keystore = &Wallet{}
