package keystore_test

import (
	"github.com/olympus-protocol/ogen/internal/keystore"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var k = keystore.NewKeystore()

func init() {
	_ = os.Mkdir(testdata.Node1Folder, 0777)
}

func TestKeystore_OpenKeystoreWithoutKeystore(t *testing.T) {
	err := k.OpenKeystore()
	assert.NotNil(t, err)
	assert.Equal(t, keystore.ErrorNotInitialized, err)
}

func TestKeystore_CreateKeystore(t *testing.T) {
	err := k.CreateKeystore()
	assert.NoError(t, err)
}

func TestKeystore_CreateKeystoreWithKeystoreOpen(t *testing.T) {
	err := k.CreateKeystore()
	assert.NotNil(t, err)
	assert.Equal(t, err, keystore.ErrorAlreadyOpen)
	assert.NoError(t, k.Close())
}

func TestKeystore_CreateKeystoreWithAlreadyExisting(t *testing.T) {
	err := k.CreateKeystore()
	assert.NotNil(t, err)
	assert.Equal(t, err, keystore.ErrorKeystoreExists)
}

func TestKeystore_OpenKeystore(t *testing.T) {
	err := k.OpenKeystore()
	assert.NoError(t, err)
	cleanFolder1()
}

func cleanFolder1() {
	_ = os.RemoveAll(testdata.Node1Folder)
}
