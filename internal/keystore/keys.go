package keystore

import (
	"bytes"
	"encoding/binary"
	"github.com/olympus-protocol/ogen/pkg/bip39"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/hdwallet"
	"go.etcd.io/bbolt"
	"strconv"
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

// HasKeysToParticipate returns true if the keystore has keys to participate
func (k *keystore) HasKeysToParticipate() bool {
	keys, err := k.GetValidatorKeys()
	if err != nil {
		return false
	}
	return len(keys) > 0
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

		err := bkt.ForEach(func(k, v []byte) error {
			key := new(Key)
			err := key.Unmarshal(v)
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

	lastPath := k.GetLastPath()
	newPath := lastPath + int(amount)
	err := k.SetLastPath(newPath)
	if err != nil {
		return nil, err
	}
	keys := make([]*Key, amount)

	seed := bip39.NewSeed(k.GetMnemonic(), "")
	for i := range keys {
		aggPath := i + 1
		// Generate a new key
		path := "m/12381/1997/0/" + strconv.Itoa(lastPath+aggPath)

		sec, err := hdwallet.CreateHDWallet(seed, path)
		if err != nil {
			return nil, err
		}

		key := &Key{
			Secret: sec,
			Enable: true,
			Path:   int64(lastPath + aggPath),
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

		kr, err := key.Marshal()
		if err != nil {
			return err
		}

		err = bkt.Put(pub.Marshal(), kr)
		if err != nil {
			return err
		}

		return nil
	})
}

// ToggleKey toggles the keystore key as enabled/disabled
func (k *keystore) ToggleKey(pub [48]byte, value bool) error {
	err := k.db.Update(func(tx *bbolt.Tx) error {
		keysBkt := tx.Bucket(keysBucket)
		k := keysBkt.Get(pub[:])
		if k == nil {
			return ErrorKeyNotOnKeystore
		}
		key := new(Key)
		err := key.Unmarshal(k)
		if err != nil {
			return err
		}
		key.Enable = value

		rawK, err := key.Marshal()
		if err != nil {
			return err
		}
		err = keysBkt.Put(pub[:], rawK)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
