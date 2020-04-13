package chain

import (
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/primitives"
)

// BlockchainNotifee is a type that is notified when something changes with the
// blockchain.
type BlockchainNotifee interface {
	NewTip(*index.BlockRow, *primitives.Block)
}

// Notify registers a notifee to be notified.
func (ch *Blockchain) Notify(n BlockchainNotifee) {
	ch.notifeeLock.Lock()
	defer ch.notifeeLock.Unlock()

	ch.notifees[n] = struct{}{}
}

// Unnotify unregisters a notifee to be notified.
func (ch *Blockchain) Unnotify(n BlockchainNotifee) {
	ch.notifeeLock.Lock()
	defer ch.notifeeLock.Unlock()

	delete(ch.notifees, n)
}
