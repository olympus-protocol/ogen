package bls_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/bls"
)

func Test_CombinedSignatureSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v bls.CombinedSignature
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc bls.CombinedSignature
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MultipubSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(32, 32)
	var v bls.Multipub
	f.Fuzz(&v)
	ser := v.Marshal()

	var desc bls.Multipub
	err := desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MultisigSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(32, 32)
	var m bls.Multipub
	var s [][96]byte
	f.Fuzz(&m)
	f.Fuzz(&s)
	v := bls.Multisig{
		PublicKey:  &m,
		Signatures: s,
		KeysSigned: []uint8{},
	}

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc bls.Multisig
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}
