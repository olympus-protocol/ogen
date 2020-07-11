package aesbls_test

import (
	"reflect"
	"testing"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/aesbls"
)

var encKey = []byte("test")

func Test_EncryptDecrypt(t *testing.T) {
	rand := bls.RandKey()
	keyBytes := rand.Marshal()
	nonce, salt, cipher, err := aesbls.Encrypt(keyBytes, encKey)
	if err != nil {
		t.Fatal(err)
	}
	key, err := aesbls.Decrypt(cipher, nonce, encKey, salt)
	if err != nil {
		t.Fatal(err)
	}
	equal := reflect.DeepEqual(key.Marshal(), keyBytes)
	if !equal {
		t.Fatal("encrypted and decrypted don't match")
	}
}
