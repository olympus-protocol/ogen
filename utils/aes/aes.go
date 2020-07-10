package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

// Decrypt uses aes decryption to decrypt an encrypted bls private key using a nonce and a salt.
func Decrypt(encryptedKey []byte, nonce []byte, key []byte, salt []byte) (*bls.SecretKey, error) {
	encryptionKey := pbkdf2.Key(key, salt, 20000, 32, sha512.New)

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, errors.Wrap(err, "error creating cipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "error creating GCM")
	}

	blsKeyBytes, err := aesgcm.Open(nil, nonce, encryptedKey, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not decrypt master key")
	}

	secKey, err := bls.SecretKeyFromBytes(blsKeyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to deserialize bls private key")
	}
	return secKey, nil
}

// Encrypt encrypts a bls private key and returns a random nonce and a salt.
func Encrypt(secret []byte, key []byte) (nonce []byte, salt [8]byte, encryptedKey []byte, err error) {
	// Generate a random salt
	_, err = rand.Reader.Read(salt[:])
	if err != nil {
		return []byte{}, [8]byte{}, nil, errors.Wrap(err, "error reading from random")
	}
	// Generate a random nonce
	encKey := pbkdf2.Key(key, salt[:], 20000, 32, sha512.New)
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return []byte{}, [8]byte{}, nil, errors.Wrap(err, "error creating cipher")
	}
	nonce = make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return []byte{}, [8]byte{}, nil, errors.Wrap(err, "error reading from random")
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, [8]byte{}, nil, errors.Wrap(err, "error creating GCM")
	}
	ciphertext := aesgcm.Seal(nil, nonce, secret, nil)
	return nonce, salt, ciphertext, nil
}
