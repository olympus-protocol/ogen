package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

func (b *Wallet) unlock(authentication []byte) (*bls.SecretKey, error) {
	encryptionKey := pbkdf2.Key(authentication, b.info.salt, 20000, 32, sha512.New)

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, errors.Wrap(err, "error creating cipher")
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(err, "error reading from random")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "error creating GCM")
	}

	masterSeed, err := aesgcm.Open(nil, b.info.nonce, b.info.encryptedMaster, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not decrypt master key")
	}

	var secretKeyBytes [32]byte
	copy(secretKeyBytes[:], masterSeed)

	secKey := bls.DeriveSecretKey(secretKeyBytes)

	return &secKey, nil
}

func (b *Wallet) unlockIfNeeded(authentication []byte) (*bls.SecretKey, error) {
	if authentication == nil || len(authentication) == 0 {
		return nil, fmt.Errorf("wallet locked, need authentication")
	}

	privVal, err := b.unlock(authentication)
	if err != nil {
		return nil, err
	}

	return privVal, nil
}

func (b *Wallet) GetAddress() string {
	return b.info.address
}
