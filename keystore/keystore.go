package keystore

import (
	"path"

	"github.com/olympus-protocol/ogen/utils/logger"
	"go.etcd.io/bbolt"
)

type Keystore struct {
	db  *bbolt.DB
	log *logger.Logger
}

// NewKeystore creates a new keystore.
func NewKeystore(pathStr string, log *logger.Logger, password string) (*Keystore, error) {
	db, err := bbolt.Open(path.Join(pathStr, "keystore.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(keysBucket); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	w := &Keystore{
		db:  db,
		log: log,
	}
	return w, nil
}

func (b *Keystore) Close() error {
	return b.db.Close()
}
