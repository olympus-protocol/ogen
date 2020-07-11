package keystore_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/olympus-protocol/ogen/keystore"
)

var pass = "test_pass"

func Test_KestoreCreate(t *testing.T) {
	k, err := createKeystore()
	if err != nil {
		t.Fatal(err)
	}
	keys, err := k.GetValidatorKeys()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 8 {
		t.Fatal("wrong number of keys")
	}
	for _, key := range keys {
		if v := k.HasValidatorKey(key.PublicKey().Marshal()); !v {
			t.Fatal("key not found")
		}
	}
	for _, key := range keys {
		secret, ok := k.GetValidatorKey(key.PublicKey().Marshal())
		if !ok {
			t.Fatal("key not found")
		}
		equal := reflect.DeepEqual(secret, key)
		if !equal {
			t.Fatal("keys don't match")
		}
	}
	clean()
}

func Test_KeystoreOpen(t *testing.T) {
	k, err := createKeystore()
	if err != nil {
		t.Fatal(err)
	}
	k.Close()
	k, err = keystore.NewKeystore("./", nil, pass)
	if err != nil {
		t.Fatal(err)
	}
	keys, err := k.GetValidatorKeys()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 8 {
		t.Fatal("wrong number of keys")
	}
	for _, key := range keys {
		if v := k.HasValidatorKey(key.PublicKey().Marshal()); !v {
			t.Fatal("key not found")
		}
	}
	for _, key := range keys {
		secret, ok := k.GetValidatorKey(key.PublicKey().Marshal())
		if !ok {
			t.Fatal("key not found")
		}
		equal := reflect.DeepEqual(secret, key)
		if !equal {
			t.Fatal("keys don't match")
		}
	}
	clean()
}

func Test_KeystoreOpenWithWrongPassword(t *testing.T) {
	k, err := createKeystore()
	if err != nil {
		t.Fatal(err)
	}
	k.Close()
	_, err = keystore.NewKeystore("./", nil, "wrong_password")
	if err == nil {
		t.Fatal("open keystore should give an error")
	}
	clean()
}

func createKeystore() (*keystore.Keystore, error) {
	k, err := keystore.NewKeystore("./", nil, pass)
	if err != nil {
		return nil, err
	}
	newKeys, err := k.GenerateNewValidatorKey(8, pass)
	if err != nil {
		return nil, err
	}
	if len(newKeys) != 8 {
		return nil, err
	}
	return k, err
}

func clean() {
	os.RemoveAll("./keystore.db")
}
