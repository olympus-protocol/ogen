package keystore_test

import (
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/pkg/bls"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var k2 = keystore.NewKeystore(testdata.Node2Folder, nil)

func init() {
	_ = os.Mkdir(testdata.Node2Folder, 0777)
	_ = k2.CreateKeystore()
}

var keys []bls_interface.SecretKey

func TestKeystore_GenerateNewValidatorKey(t *testing.T) {
	var err error
	keys, err = k2.GenerateNewValidatorKey(8)
	assert.NoError(t, err)
}

func TestKeystore_GetValidatorKeys(t *testing.T) {
	newKeys, err := k2.GetValidatorKeys()
	assert.NoError(t, err)
	assert.Equal(t, len(keys), len(newKeys))
}

func TestKeystore_GetValidatorKey(t *testing.T) {
	var pub [48]byte
	copy(pub[:], keys[0].PublicKey().Marshal())
	key, ok := k2.GetValidatorKey(pub)
	assert.True(t, ok)
	assert.Equal(t, keys[0], key)
}

func TestKeystore_GetValidatorKeyWithoutKey(t *testing.T) {
	var pub [48]byte
	copy(pub[:], bls.CurrImplementation.RandKey().PublicKey().Marshal())
	_, ok := k2.GetValidatorKey(pub)
	assert.False(t, ok)
	cleanFolder2()
}

func cleanFolder2() {
	_ = os.RemoveAll(testdata.Node2Folder)
}
