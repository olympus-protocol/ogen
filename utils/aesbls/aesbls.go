package aesbls

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
func Decrypt(nonce [12]byte, salt [8]byte, encryptedKey []byte, key []byte) (*bls.SecretKey, error) {
	encryptionKey := pbkdf2.Key(key, salt[:], 20000, 32, sha512.New)

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, errors.Wrap(err, "error creating cipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "error creating GCM")
	}

	blsKeyBytes, err := aesgcm.Open(nil, nonce[:], encryptedKey, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not decrypt master key")
	}

	secKey, err := bls.SecretKeyFromBytes(blsKeyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to deserialize bls private key")
	}
	return secKey, nil
}

// SimpleEncrypt encrypts a bls private key with a known nonce and salt.
func SimpleEncrypt(secret []byte, key []byte, nonce [12]byte, salt [8]byte) (encryptedKey []byte, err error) {
	encKey := pbkdf2.Key(key, salt[:], 20000, 32, sha512.New)

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, errors.Wrap(err, "error creating cipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "error creating GCM")
	}

	return aesgcm.Seal(nil, nonce[:], secret, nil), nil
}

// Encrypt encrypts a bls private key and returns a random nonce and a salt.
func Encrypt(secret []byte, key []byte) (nonce [12]byte, salt [8]byte, encryptedKey []byte, err error) {
	salt, err = RandSalt()
	if err != nil {
		return [12]byte{}, [8]byte{}, nil, err
	}
	encKey := pbkdf2.Key(key, salt[:], 20000, 32, sha512.New)
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return [12]byte{}, [8]byte{}, nil, errors.Wrap(err, "error creating cipher")
	}
	nonce, err = RandNonce()
	if err != nil {
		return [12]byte{}, [8]byte{}, nil, err
	}
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return [12]byte{}, [8]byte{}, nil, errors.Wrap(err, "error reading from random")
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return [12]byte{}, [8]byte{}, nil, errors.Wrap(err, "error creating GCM")
	}
	return nonce, salt, aesgcm.Seal(nil, nonce[:], secret, nil), nil
}

// RandSalt generates a random salt from random reader.
func RandSalt() ([8]byte, error) {
	var salt [8]byte
	_, err := rand.Read(salt[:])
	if err != nil {
		return [8]byte{}, err
	}
	return salt, nil
}

// RandNonce generates a random nonce from random reader.
func RandNonce() ([12]byte, error) {
	var nonce [12]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		return [12]byte{}, err
	}
	return nonce, nil
}
