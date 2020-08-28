package aesbls_test

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"

	"github.com/olympus-protocol/ogen/pkg/aesbls"
	"github.com/olympus-protocol/ogen/pkg/bls"
)

var encKey = []byte("test")
var encKeyWrong = []byte("test2")

func Test_EncryptDecrypt(t *testing.T) {

	rand, err := bls.RandKey()
	assert.NoError(t, err)

	keyBytes := rand.Marshal()
	nonce, salt, cipher, err := aesbls.Encrypt(keyBytes, encKey)
	assert.NoError(t, err)

	key, err := aesbls.Decrypt(nonce, salt, cipher, encKey)
	assert.NoError(t, err)

	equal := reflect.DeepEqual(key.Marshal(), keyBytes)
	assert.True(t, equal)

	wrongKey, err := aesbls.Decrypt(nonce, salt, cipher, encKeyWrong)
	assert.Equal(t, aesbls.ErrorDecrypt, err)

	assert.Nil(t, wrongKey)
}

func Test_SimpleEncryptDecrypt(t *testing.T) {
	rand, err := bls.RandKey()
	assert.NoError(t, err)

	keyBytes := rand.Marshal()

	nonce := aesbls.RandNonce()

	salt := aesbls.RandSalt()

	cipher, err := aesbls.SimpleEncrypt(keyBytes, encKey, nonce, salt)
	assert.NoError(t, err)

	key, err := aesbls.Decrypt(nonce, salt, cipher, encKey)
	assert.NoError(t, err)

	equal := reflect.DeepEqual(key.Marshal(), keyBytes)
	assert.True(t, equal)

	wrongKey, err := aesbls.Decrypt(nonce, salt, cipher, encKeyWrong)
	assert.Equal(t, aesbls.ErrorDecrypt, err)

	assert.Nil(t, wrongKey)
}
