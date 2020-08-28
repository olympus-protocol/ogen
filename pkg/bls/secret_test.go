package bls_test

import (
	"encoding/hex"
	"github.com/olympus-protocol/ogen/pkg/bls"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)


func TestRandKey(t *testing.T) {
	for i := 0; i < 50000; i++ {
		r, err := bls.RandKey()
		assert.NoError(t, err)
		mar := r.Marshal()
		_, err = bls.SecretKeyFromBytes(mar)
		assert.NoError(t, err)
	}
}

func TestRandKeyWithSign(t *testing.T) {
	for i := 0; i < 2000; i++ {
		r, err := bls.RandKey()
		if err != nil {
			assert.NoError(t, err)
		}
		sig := r.Sign([]byte("a"))
		assert.True(t, sig.Verify(r.PublicKey(), []byte("a")))
		mar := r.Marshal()
		_, err = bls.SecretKeyFromBytes(mar)
		assert.NoError(t, err)

	}
}

func TestSecretKey_ToWIF(t *testing.T) {

	bls.Initialize(testdata.TestParams)

	secBytes, err := hex.DecodeString("28291cbbfaba8ca4d350a7a7f59cac06f7cb2a346e396389d30b0a6c5b59ec73")
	assert.NoError(t, err)

	sec, err := bls.SecretKeyFromBytes(secBytes)
	assert.NoError(t, err)

	wif := sec.ToWIF()
	assert.Equal(t, "itprv19q53ewl6h2x2f56s57nlt89vqmmuk235dcuk8zwnpv9xck6ea3esywcazn", wif)
}

func TestMarshalUnmarshal(t *testing.T) {
	k, err := bls.RandKey()
	assert.NoError(t, err)
	b := k.Marshal()
	var b32 [32]byte
	copy(b32[:], b)
	pk, err := bls.SecretKeyFromBytes(b32[:])
	require.NoError(t, err)
	pk2, err := bls.SecretKeyFromBytes(b32[:])
	require.NoError(t, err)
	assert.Equal(t, pk.Marshal(), pk2.Marshal())
}

func TestSecretKeyFromBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		err   error
	}{
		{
			name: "Nil",
			err:  bls.ErrorSecUnmarshal,
		},
		{
			name:  "Empty",
			input: []byte{},
			err:   bls.ErrorSecUnmarshal,
		},
		{
			name:  "Short",
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			err:   bls.ErrorSecUnmarshal,
		},
		{
			name:  "Long",
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			err:   bls.ErrorSecUnmarshal,
		},
		{
			name:  "Bad",
			input: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			err:   bls.ErrorSecUnmarshal,
		},
		{
			name:  "Good",
			input: []byte{0x25, 0x29, 0x5f, 0x0d, 0x1d, 0x59, 0x2a, 0x90, 0xb3, 0x33, 0xe2, 0x6e, 0x85, 0x14, 0x97, 0x08, 0x20, 0x8e, 0x9f, 0x8e, 0x8b, 0xc1, 0x8f, 0x6c, 0x77, 0xbd, 0x62, 0xf8, 0xad, 0x7a, 0x68, 0x66},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := bls.SecretKeyFromBytes(test.input)
			if test.err != nil {
				assert.Equal(t, test.err, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.input, res.Marshal())
			}
		})
	}
}

func TestSerialize(t *testing.T) {

	rk, err := bls.RandKey()
	assert.NoError(t, err)
	b := rk.Marshal()

	_, err = bls.SecretKeyFromBytes(b)
	assert.NoError(t, err)
}
