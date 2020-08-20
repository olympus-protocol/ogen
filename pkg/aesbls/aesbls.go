package aesbls

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"github.com/olympus-protocol/ogen/pkg/bls"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

var (
	ErrorCipher      = errors.New("error creating the encryption cypher")
	ErrorGCM         = errors.New("error loading the gcm cypher, nonce or salt information may be wrong")
	ErrorDecrypt     = errors.New("unable to decrypt key with key provided")
	ErrorDeserialize = errors.New("unable to deserialize the bls encoded key")
)

// Decrypt uses aes decryption to decrypt an encrypted bls private key using a nonce and a salt.
func Decrypt(nonce [12]byte, salt [8]byte, encryptedKey []byte, key []byte) (bls_interface.SecretKey, error) {
	encryptionKey := pbkdf2.Key(key, salt[:], 20000, 32, sha512.New)

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, ErrorCipher
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrorGCM
	}

	blsKeyBytes, err := aesgcm.Open(nil, nonce[:], encryptedKey, nil)
	if err != nil {
		return nil, ErrorDecrypt
	}

	secKey, err := bls.CurrImplementation.SecretKeyFromBytes(blsKeyBytes)
	if err != nil {
		return nil, ErrorDeserialize
	}
	return secKey, nil
}

// SimpleEncrypt encrypts a bls private key with a known nonce and salt.
func SimpleEncrypt(secret []byte, key []byte, nonce [12]byte, salt [8]byte) (encryptedKey []byte, err error) {
	encKey := pbkdf2.Key(key, salt[:], 20000, 32, sha512.New)

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, ErrorCipher
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrorGCM
	}

	return aesgcm.Seal(nil, nonce[:], secret, nil), nil
}

// Encrypt encrypts a bls private key and returns a random nonce and a salt.
func Encrypt(secret []byte, key []byte) (nonce [12]byte, salt [8]byte, encryptedKey []byte, err error) {
	salt = RandSalt()

	encKey := pbkdf2.Key(key, salt[:], 20000, 32, sha512.New)

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return [12]byte{}, [8]byte{}, nil, ErrorCipher
	}

	nonce = RandNonce()

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return [12]byte{}, [8]byte{}, nil, ErrorGCM
	}

	return nonce, salt, aesgcm.Seal(nil, nonce[:], secret, nil), nil
}

// RandSalt generates a random salt from random reader.
func RandSalt() [8]byte {
	var salt [8]byte
	_, _ = rand.Read(salt[:])
	return salt
}

// RandNonce generates a random nonce from random reader.
func RandNonce() [12]byte {
	var nonce [12]byte
	_, _ = rand.Read(nonce[:])
	return nonce
}
