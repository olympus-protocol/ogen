package keystore

import (
	"bytes"
	"encoding/binary"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"go.etcd.io/bbolt"
)

type Key struct {
	Secret common.SecretKey
	Enable bool
	Path   int64
}

func (k *Key) Marshal() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	var sec [32]byte
	copy(sec[:], k.Secret.Marshal())

	err := binary.Write(buf, binary.LittleEndian, sec)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, k.Enable)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, k.Path)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (k *Key) Unmarshal(b []byte) error {
	buf := bytes.NewBuffer(b)

	var sec [32]byte
	err := binary.Read(buf, binary.LittleEndian, &sec)
	if err != nil {
		return err
	}

	blsSec, err := bls.SecretKeyFromBytes(sec[:])
	if err != nil {
		return err
	}

	k.Secret = blsSec

	err = binary.Read(buf, binary.LittleEndian, &k.Enable)
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &k.Path)
	if err != nil {
		return err
	}

	return nil
}

// GetValidatorKey returns the private key from the specified public key or false if doesn't exists.
func (k *keystore) GetValidatorKey(pubkey [48]byte) (*Key, bool) {

	if !k.open {
		return nil, false
	}

	var key []byte
	err := k.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket(keysBucket)
		key = bkt.Get(pubkey[:])
		return nil
	})
	if err != nil {
		return nil, false
	}

	if key == nil {
		return nil, false
	}

	keystoreKey := new(Key)
	err = keystoreKey.Unmarshal(key)
	if err != nil {
		return nil, false
	}

	return keystoreKey, true
}

// GetValidatorKeys returns all keys on keystore.
func (k *keystore) GetValidatorKeys() ([]*Key, error) {

	if !k.open {
		return nil, ErrorNoOpen
	}

	var keys []*Key

	err := k.db.View(func(tx *bbolt.Tx) error {

		bkt := tx.Bucket(keysBucket)

		err := bkt.ForEach(func(keypub, keyprv []byte) error {
			key := new(Key)
			err := key.Unmarshal(keyprv)
			if err != nil {
				return err
			}

			keys = append(keys, key)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// GenerateNewValidatorKey generates new validator keys and adds it to the map and database.
func (k *keystore) GenerateNewValidatorKey(amount uint64) ([]*Key, error) {
	if !k.open {
		return nil, ErrorNoOpen
	}

	keys := make([]*Key, amount)

	for i := range keys {
		// Generate a new key
		sec, err := bls.RandKey()
		if err != nil {
			return nil, err
		}

		key := &Key{
			Secret: sec,
			Enable: true,
		}

		err = k.AddKey(key)
		if err != nil {
			return nil, err
		}
		keys[i] = key
	}

	return keys, nil
}

func (k *keystore) AddKey(key *Key) error {
	if !k.open {
		return ErrorNoOpen
	}

	return k.db.Update(func(tx *bbolt.Tx) error {

		pub := key.Secret.PublicKey()

		bkt := tx.Bucket(keysBucket)

		err := bkt.Put(pub.Marshal(), key.Secret.Marshal())
		if err != nil {
			return err
		}

		return nil
	})
}
