package bls_test

import (
	"testing"

	"github.com/olympus-protocol/ogen/bls"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
)

func Test_CombinedSignatureSerialize(t *testing.T) {

	ser, err := testdata.CombinedSignature.Marshal()
	
	assert.NoError(t, err)

	var desc bls.CombinedSignature

	err = desc.Unmarshal(ser)
	
	assert.NoError(t, err)

	assert.Equal(t, testdata.CombinedSignature, desc)
}

func Test_MultipubSerialize(t *testing.T) {
	ser := testdata.Multipub.Marshal()

	var desc bls.Multipub

	err := desc.Unmarshal(ser)
	
	assert.NoError(t, err)

	assert.Equal(t, testdata.Multipub, desc)
}

func Test_MultisigSerialize(t *testing.T) {
	ser, err := testdata.Multisig.Marshal()
	
	assert.NoError(t, err)

	var desc bls.Multisig

	err = desc.Unmarshal(ser)
	
	assert.NoError(t, err)

	assert.Equal(t, testdata.Multisig, desc)
}
