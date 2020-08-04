package bls_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
)

func TestCorrectnessMultisig(t *testing.T) {
	secretKeys := make([]*bls.SecretKey, 20)
	publicKeys := make([]*bls.PublicKey, 20)

	for i := range secretKeys {
		secretKeys[i] = bls.RandKey()
		publicKeys[i] = secretKeys[i].PublicKey()
	}

	// create 10-of-20 multipub
	multiPub := bls.NewMultipub(publicKeys, 10)
	multisig := bls.NewMultisig(multiPub)

	msg := []byte("hello there!")

	for i := 0; i < 9; i++ {
		assert.NoError(t, multisig.Sign(secretKeys[i], msg))
	}

	assert.False(t, multisig.Verify(msg))

	assert.NoError(t, multisig.Sign(secretKeys[9], msg))

	assert.True(t, multisig.Verify(msg))

	for i := 10; i < 20; i++ {
		assert.NoError(t, multisig.Sign(secretKeys[i], msg))
	}

	assert.True(t, multisig.Verify(msg))

	multiPub.ToBech32(params.AccountPrefixes{
		Multisig: "olmul",
	})
}

func TestMultisigSerializeSign(t *testing.T) {
	secretKeys := make([]*bls.SecretKey, 20)
	publicKeys := make([]*bls.PublicKey, 20)

	for i := range secretKeys {
		secretKeys[i] = bls.RandKey()
		publicKeys[i] = secretKeys[i].PublicKey()
	}

	// create 10-of-20 multipub
	multiPub := bls.NewMultipub(publicKeys, 10)
	multisig := bls.NewMultisig(multiPub)

	msg := []byte("hello there!")

	for i := 0; i < 10; i++ {
		assert.NoError(t, multisig.Sign(secretKeys[i], msg))
	}

	multiBytes, err := multisig.Marshal()

	assert.NoError(t, err)

	newMulti := new(bls.Multisig)

	assert.NoError(t, newMulti.Unmarshal(multiBytes))

	assert.True(t, newMulti.Verify(msg))
}
