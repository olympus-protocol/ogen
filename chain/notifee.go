package chain

import (
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/primitives"
)

// BlockchainNotifee is a type that is notified when something changes with the
// blockchain.
type BlockchainNotifee interface {
	// NewTip notifies of a new tip added to the blockchain. Do not mutate state.
	NewTip(*index.BlockRow, *primitives.Block, *primitives.State)
	ProposerSlashingConditionViolated(slashing primitives.ProposerSlashing)
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
