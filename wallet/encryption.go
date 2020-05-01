package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

func (b *Wallet) unlock(authentication []byte) error {
	if !atomic.CompareAndSwapUint32(&b.masterLock, 0, 1) {
		return nil
	}

	encryptionKey := pbkdf2.Key(authentication, b.info.salt, 20000, 32, sha512.New)

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return errors.Wrap(err, "error creating cipher")
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return errors.Wrap(err, "error reading from random")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return errors.Wrap(err, "error creating GCM")
	}

	masterSeed, err := aesgcm.Open(nil, b.info.nonce, b.info.encryptedMaster, nil)
	if err != nil {
		return errors.Wrap(err, "could not decrypt master key")
	}

	var secretKeyBytes [32]byte
	copy(secretKeyBytes[:], masterSeed)

	secKey := bls.DeriveSecretKey(secretKeyBytes)

	b.masterPriv.Store(&secKey)

	go func() {
		<-time.After(time.Minute * 2)
		b.masterPriv.Store(nil)
		atomic.StoreUint32(&b.masterLock, 0)
	}()

	return nil
}

func (b *Wallet) unlockIfNeeded(authentication []byte) (*bls.SecretKey, error) {
	privVal := b.masterPriv.Load()
	if privVal == nil {
		if authentication == nil || len(authentication) == 0 {
			return nil, fmt.Errorf("wallet locked, need authentication")
		}

		if err := b.unlock(authentication); err != nil {
			return nil, err
		}

		privVal = b.masterPriv.Load()
	}

	return privVal.(*bls.SecretKey), nil
}

func (b *Wallet) GetAddress() string {
	return b.info.address
}
