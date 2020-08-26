package multisig_test

import (
	"github.com/olympus-protocol/ogen/pkg/bls/multisig"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/pkg/bls"
)

func TestCorrectnessMultisig(t *testing.T) {
	secretKeys := make([]*bls.SecretKey, 20)
	publicKeys := make([]*bls.PublicKey, 20)

	for i := range secretKeys {
		secretKeys[i] = bls.RandKey()
		publicKeys[i] = secretKeys[i].PublicKey()
	}

	// create 10-of-20 multipub
	multiPub := multisig.NewMultipub(publicKeys, 10)

	ser, err := multiPub.Marshal()
	assert.NoError(t, err)

	desc := new(multisig.Multipub)

	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	ms := multisig.NewMultisig(multiPub)

	msg := []byte("hello there!")

	for i := 0; i < 9; i++ {
		assert.NoError(t, ms.Sign(secretKeys[i], msg))
	}

	assert.False(t, ms.Verify(msg))

	assert.NoError(t, ms.Sign(secretKeys[9], msg))

	assert.True(t, ms.Verify(msg))

	for i := 10; i < 20; i++ {
		assert.NoError(t, ms.Sign(secretKeys[i], msg))
	}

	assert.True(t, ms.Verify(msg))

	//_, err = multiPub.ToBech32()
	//assert.NoError(t, err)
}

func TestMultisigSerializeSign(t *testing.T) {
	secretKeys := make([]*bls.SecretKey, 20)
	publicKeys := make([]*bls.PublicKey, 20)

	for i := range secretKeys {
		secretKeys[i] = bls.RandKey()
		publicKeys[i] = secretKeys[i].PublicKey()
	}

	// create 10-of-20 multipub
	multiPub := multisig.NewMultipub(publicKeys, 10)
	ms := multisig.NewMultisig(multiPub)

	msg := []byte("hello there!")

	for i := 0; i < 10; i++ {
		assert.NoError(t, ms.Sign(secretKeys[i], msg))
	}

	multiBytes, err := ms.Marshal()

	assert.NoError(t, err)

	newMulti := new(multisig.Multisig)

	assert.NoError(t, newMulti.Unmarshal(multiBytes))

	assert.True(t, newMulti.Verify(msg))
}

func TestMultipubCopy(t *testing.T) {
	secretKeys := make([]*bls.SecretKey, 20)
	publicKeys := make([]*bls.PublicKey, 20)
	for i := range secretKeys {
		secretKeys[i] = bls.RandKey()
		publicKeys[i] = secretKeys[i].PublicKey()
	}

	mp := multisig.NewMultipub(publicKeys, 10)

	cp := mp.Copy()

	mp.NumNeeded = 1
	assert.Equal(t, uint64(10), cp.NumNeeded)

	mp.PublicKeys = nil
	assert.Equal(t, 20, len(cp.PublicKeys))
}

func TestMultisigCopy(t *testing.T) {
	secretKeys := make([]*bls.SecretKey, 20)
	publicKeys := make([]*bls.PublicKey, 20)
	for i := range secretKeys {
		secretKeys[i] = bls.RandKey()
		publicKeys[i] = secretKeys[i].PublicKey()
	}

	mp := multisig.NewMultipub(publicKeys, 10)
	ms := multisig.NewMultisig(mp)

	msg := []byte("hello there!")

	for i := 0; i < 10; i++ {
		assert.NoError(t, ms.Sign(secretKeys[i], msg))
	}

	cp := ms.Copy()

	ms.KeysSigned.Set(11)
	assert.False(t, cp.KeysSigned.Get(11))

	ms.Signatures = nil
	assert.Equal(t, 10, len(cp.Signatures))

	ms.PublicKey = nil
	assert.NotNil(t, cp.PublicKey)
}
