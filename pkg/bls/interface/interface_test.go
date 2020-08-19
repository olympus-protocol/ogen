package bls_interface_test

import (
	"github.com/olympus-protocol/ogen/pkg/bls/bls_blst"
	"github.com/olympus-protocol/ogen/pkg/bls/bls_herumi"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

type TestVectors struct {
	Pairs []Pair `json:"pairs"`
}

type Pair struct {
	Private   []byte
	Public    []byte
	Msg       []byte
	Signature []byte
}

var testVectors TestVectors

func init() {
	impl := bls_herumi.HerumiImplementation{}
	var pairs []Pair
	for i := 0; i < 100; i++ {
		key := impl.RandKey()
		msg := "test-" + strconv.Itoa(i)
		sig := key.Sign([]byte(msg))
		pairs = append(pairs, Pair{
			Private:   key.Marshal(),
			Public:    key.PublicKey().Marshal(),
			Msg:       []byte("test-" + strconv.Itoa(i)),
			Signature: sig.Marshal(),
		})
	}
	testVectors = TestVectors{
		Pairs: pairs,
	}
}

func TestHerumi(t *testing.T) {
	impl := bls_herumi.HerumiImplementation{}

	for _, pair := range testVectors.Pairs {
		k, err := impl.SecretKeyFromBytes(pair.Private)
		assert.NoError(t, err)
		p, err := impl.PublicKeyFromBytes(pair.Public)
		assert.NoError(t, err)
		s, err := impl.SignatureFromBytes(pair.Signature)
		assert.NoError(t, err)

		assert.Equal(t, k.PublicKey().Marshal(), pair.Public)
		assert.Equal(t, p.Marshal(), pair.Public)
		assert.Equal(t, s.Marshal(), pair.Signature)

		newSig := k.Sign(pair.Msg)

		assert.Equal(t, newSig.Marshal(), pair.Signature)

		assert.True(t, newSig.Verify(p, pair.Msg))
	}
}

func TestBlst(t *testing.T) {
	impl := bls_blst.BlstImplementation{}

	for _, pair := range testVectors.Pairs {
		k, err := impl.SecretKeyFromBytes(pair.Private)
		assert.NoError(t, err)
		p, err := impl.PublicKeyFromBytes(pair.Public)
		assert.NoError(t, err)
		s, err := impl.SignatureFromBytes(pair.Signature)
		assert.NoError(t, err)

		assert.Equal(t, k.PublicKey().Marshal(), pair.Public)
		assert.Equal(t, p.Marshal(), pair.Public)
		assert.Equal(t, s.Marshal(), pair.Signature)

		newSig := k.Sign(pair.Msg)

		assert.Equal(t, newSig.Marshal(), pair.Signature)

		assert.True(t, newSig.Verify(p, pair.Msg))
	}
}
