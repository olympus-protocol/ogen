package wallet

import (
	"context"
	"sync"

	"github.com/olympus-protocol/ogen/mempool"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

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

	chain         *chain.Blockchain
	mempool       *mempool.CoinsMempool
	actionMempool *mempool.ActionMempool

	txTopic      *pubsub.Topic
	depositTopic *pubsub.Topic

	hasMaster bool

	info          walletInfo
	lastNonceLock sync.Mutex
	ctx           context.Context

	*ValidatorWallet
}

var walletDBKey = []byte("encryption-key-ciphertext")
var walletDBSalt = []byte("encryption-key-salt")
var walletDBNonce = []byte("encryption-key-nonce")
var walletDBLastTxNonce = []byte("last-tx-nonce")
var walletDBAddress = []byte("wallet-address")

// NewWallet creates a new wallet.
func NewWallet(ctx context.Context, c Config, params params.ChainParams, ch *chain.Blockchain, hostnode *peers.HostNode, walletDB *badger.DB, mempool *mempool.CoinsMempool, actionMempool *mempool.ActionMempool) (*Wallet, error) {
	txTopic, err := hostnode.Topic("tx")
	if err != nil {
		return nil, err
	}

	depositTopic, err := hostnode.Topic("deposits")
	if err != nil {
		return nil, err
	}

	w := &Wallet{
		db:              walletDB,
		hasMaster:       false,
		log:             c.Log,
		params:          &params,
		chain:           ch,
		txTopic:         txTopic,
		depositTopic:    depositTopic,
		mempool:         mempool,
		ctx:             ctx,
		actionMempool:   actionMempool,
		ValidatorWallet: NewValidatorWallet(walletDB),
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
