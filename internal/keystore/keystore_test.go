package keystore_test

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	config.GlobalFlags = &config.Flags{
		DataPath: "./",
	}
}

func Test_Keystore(t *testing.T) {

	ks := keystore.NewKeystore()

	err := ks.CreateKeystore()
	assert.NoError(t, err)

	mnemonic := ks.GetMnemonic()

	assert.Equal(t, 0, ks.GetLastPath())

	err = ks.Close()
	assert.NoError(t, err)

	err = ks.OpenKeystore()
	assert.NoError(t, err)

	openMnemonic := ks.GetMnemonic()
	assert.Equal(t, mnemonic, openMnemonic)

	assert.False(t, ks.HasKeysToParticipate())

	_, err = ks.GenerateNewValidatorKey(5)
	assert.NoError(t, err)

	assert.Equal(t, 5, ks.GetLastPath())

	_, err = ks.GenerateNewValidatorKey(20)
	assert.NoError(t, err)

	assert.Equal(t, 25, ks.GetLastPath())

	err = ks.Close()
	assert.NoError(t, err)

	err = ks.OpenKeystore()
	assert.NoError(t, err)

	assert.Equal(t, 25, ks.GetLastPath())

	allKeys, err := ks.GetValidatorKeys()
	assert.NoError(t, err)

	assert.Equal(t, len(allKeys), 25)

	assert.True(t, ks.HasKeysToParticipate())

	var pub [48]byte
	copy(pub[:], allKeys[0].Secret.PublicKey().Marshal())
	err = ks.ToggleKey(pub, true)
	assert.NoError(t, err)

	keyEnabled, ok := ks.GetValidatorKey(pub)
	assert.True(t, ok)
	assert.True(t, keyEnabled.Enable)

	err = ks.ToggleKey(pub, false)
	assert.NoError(t, err)

	keyDisabled, ok := ks.GetValidatorKey(pub)
	assert.True(t, ok)
	assert.False(t, keyDisabled.Enable)

}
