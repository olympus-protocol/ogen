package keystore

import (
	"crypto/rand"
	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/bls"
)

// Keystore is an interface to a simple keystore.
type Keystore interface {
	GenerateNewKey() (*bls.SecretKey, error)
	GetKey(*bls.PublicKey) (*bls.SecretKey, error)
	HasKey(*bls.PublicKey) (bool, error)
	GetKeys() ([]*bls.SecretKey, error)
	Close() error
}

type BadgerKeystore struct {
	db *badger.DB
}

func NewBadgerKeystore(path string) (*BadgerKeystore, error) {
	bdb, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	return &BadgerKeystore{
		db: bdb,
	}, nil
}

func (b *BadgerKeystore) GetKeys() ([]*bls.SecretKey, error) {
	secKeys := make([]*bls.SecretKey, 0)
	err := b.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			i := iter.Item()
			val, err := i.ValueCopy(nil)
			if err != nil {
				return err
			}
			if len(val) == 32 {
				var valBytes [32]byte
				copy(valBytes[:], val)
				secretKey := bls.DeserializeSecretKey(valBytes)
				secKeys = append(secKeys, &secretKey)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return secKeys, nil
}

func (b *BadgerKeystore) GetKey(key *bls.PublicKey) (*bls.SecretKey, error) {
	pubBytes := key.Serialize()

	var secretBytes [32]byte
	err := b.db.View(func(txn *badger.Txn) error {
		i, err := txn.Get(pubBytes[:])
		if err != nil {
			return err
		}

		_, err = i.ValueCopy(secretBytes[:])
		return err
	})
	if err != nil {
		return nil, err
	}

	secretKey := bls.DeserializeSecretKey(secretBytes)
	return &secretKey, nil
}

func (b *BadgerKeystore) HasKey(key *bls.PublicKey) (result bool, err error) {
	pubBytes := key.Serialize()

	err = b.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(pubBytes[:])
		if err == badger.ErrKeyNotFound {
			result = false
		}
		if err != nil {
			return err
		}
		result = true
		return nil
	})

	return result, err
}

func (b *BadgerKeystore) GenerateNewKey() (*bls.SecretKey, error) {
	key, err := bls.RandSecretKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	keyBytes := key.Serialize()

	pub := key.DerivePublicKey()
	pubBytes := pub.Serialize()

	err = b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(pubBytes[:], keyBytes[:])
	})

	return key, err
}

func (b *BadgerKeystore) Close() error {
	return b.db.Close()
}

var _ Keystore = &BadgerKeystore{}