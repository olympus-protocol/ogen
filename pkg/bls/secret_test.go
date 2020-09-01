package bls_test

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	b := bls.RandKey().Marshal()
	pk, err := bls.SecretKeyFromBytes(b)
	require.NoError(t, err)
	pk2, err := bls.SecretKeyFromBytes(b)
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
			err:  bls.ErrorSecSize,
		},
		{
			name:  "Empty",
			input: []byte{},
			err:   bls.ErrorSecSize,
		},
		{
			name:  "Short",
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			err:   bls.ErrorSecSize,
		},
		{
			name:  "Long",
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			err:   bls.ErrorSecSize,
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
				assert.Error(t, test.err, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.input, res.Marshal())
			}
		})
	}
}

func TestSerialize(t *testing.T) {
	rk := bls.RandKey()
	b := rk.Marshal()

	_, err := bls.SecretKeyFromBytes(b)
	assert.NoError(t, err)
}

func TestSecretKey_ToWIF(t *testing.T) {
	key := []byte{0x25, 0x29, 0x5f, 0x0d, 0x1d, 0x59, 0x2a, 0x90, 0xb3, 0x33, 0xe2, 0x6e, 0x85, 0x14, 0x97, 0x08, 0x20, 0x8e, 0x9f, 0x8e, 0x8b, 0xc1, 0x8f, 0x6c, 0x77, 0xbd, 0x62, 0xf8, 0xad, 0x7a, 0x68, 0x66}
	sec, err := bls.SecretKeyFromBytes(key)
	assert.NoError(t, err)

	assert.Equal(t, "itprv1y5547rgaty4fpvenufhg29yhpqsga8uw30qc7mrhh4303tt6dpnq7tvqcr", sec.ToWIF())
}