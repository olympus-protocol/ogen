package hdwallet_test

import (
	"github.com/olympus-protocol/ogen/pkg/bip39"
	"github.com/olympus-protocol/ogen/pkg/hdwallet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateHDWallet(t *testing.T) {
	for i := 0; i < 100; i++ {
		entropy, err := bip39.NewEntropy(256)
		assert.NoError(t, err)

		mnemonic, err := bip39.NewMnemonic(entropy)

		assert.NoError(t, err)

		seed := bip39.NewSeed(mnemonic, "password")
		key, err := hdwallet.CreateBLSHDWallet(seed, "m/1/1")
		assert.NoError(t, err)

		msg := []byte("test msg")
		sig := key.Sign(msg)
		assert.True(t, sig.Verify(key.PublicKey(), msg))
	}
}
