package keystore

import (
	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/utils/logger"
)

type Keystore struct {
	db  *badger.DB
	log *logger.Logger
}

// NewKeystore creates a new keystore.
func NewKeystore(db *badger.DB, log *logger.Logger) (*Keystore, error) {
	w := &Keystore{
		db:  db,
		log: log,
	}
	return w, nil
}

func (b *Keystore) Close() error {
	return b.db.Close()
}
