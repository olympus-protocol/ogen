package multisig_test

//
//import (
//	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
//	"github.com/olympus-protocol/ogen/pkg/bls/multisig"
//	"github.com/stretchr/testify/assert"
//	"testing"
//
//	"github.com/olympus-protocol/ogen/pkg/bls"
//)
//
//func TestCorrectnessMultisig(t *testing.T) {
//	secretKeys := make([]bls_interface.SecretKey, 20)
//	publicKeys := make([]bls_interface.PublicKey, 20)
//
//	for i := range secretKeys {
//		secretKeys[i] = bls.CurrImplementation.RandKey()
//		publicKeys[i] = secretKeys[i].PublicKey()
//	}
//
//	// create 10-of-20 multipub
//	multiPub := multisig.NewMultipub(publicKeys, 10)
//	ms := multisig.NewMultisig(multiPub)
//
//	msg := []byte("hello there!")
//
//	for i := 0; i < 9; i++ {
//		assert.NoError(t, ms.Sign(secretKeys[i], msg))
//	}
//
//	assert.False(t, ms.Verify(msg))
//
//	assert.NoError(t, ms.Sign(secretKeys[9], msg))
//
//	assert.True(t, ms.Verify(msg))
//
//	for i := 10; i < 20; i++ {
//		assert.NoError(t, ms.Sign(secretKeys[i], msg))
//	}
//
//	assert.True(t, ms.Verify(msg))
//
//	//_, err := multiPub.ToBech32()
//	//assert.NoError(t, err)
//}
//
//func TestMultisigSerializeSign(t *testing.T) {
//	secretKeys := make([]bls_interface.SecretKey, 20)
//	publicKeys := make([]bls_interface.PublicKey, 20)
//
//	for i := range secretKeys {
//		secretKeys[i] = bls.CurrImplementation.RandKey()
//		publicKeys[i] = secretKeys[i].PublicKey()
//	}
//
//	// create 10-of-20 multipub
//	multiPub := multisig.NewMultipub(publicKeys, 10)
//	ms := multisig.NewMultisig(multiPub)
//
//	msg := []byte("hello there!")
//
//	for i := 0; i < 10; i++ {
//		assert.NoError(t, ms.Sign(secretKeys[i], msg))
//	}
//
//	multiBytes, err := ms.Marshal()
//
//	assert.NoError(t, err)
//
//	newMulti := new(multisig.Multisig)
//
//	assert.NoError(t, newMulti.Unmarshal(multiBytes))
//
//	assert.True(t, newMulti.Verify(msg))
//}
