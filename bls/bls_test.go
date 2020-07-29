package bls_test

import (
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/stretchr/testify/assert"
)

var f = fuzz.New()

func Test_CombinedSignatureSerialize(t *testing.T) {
	v := new(bls.CombinedSignature)
	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(bls.CombinedSignature)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MultipubSerialize(t *testing.T) {
	v := new(bls.Multipub)
	f.Fuzz(v)

	ser := v.Marshal()

	desc := new(bls.Multipub)

	err := desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MultisigSerialize(t *testing.T) {
	v := new(bls.Multisig)
	f.Fuzz(v)

	ser, err := v.Marshal()

	assert.NoError(t, err)

	desc := new(bls.Multisig)

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}
