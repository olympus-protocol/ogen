package primitives_test

import (
	"encoding/hex"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeposit(t *testing.T) {
	v := testdata.FuzzDeposit(10, true)
	for _, c := range v {
		ser, err := c.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.Deposit)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)
	}

	nildata := testdata.FuzzDeposit(10, false)

	for _, c := range nildata {
		assert.NotPanics(t, func() {
			data, err := c.Marshal()
			assert.NoError(t, err)

			n := new(primitives.Deposit)
			err = n.Unmarshal(data)
			assert.NoError(t, err)

			assert.Equal(t, c, n)

			assert.Equal(t, [48]byte{}, n.Data.PublicKey)
			assert.Equal(t, [96]byte{}, n.Data.ProofOfPossession)

		})
	}
	d := primitives.Deposit{
		Data: &primitives.DepositData{
			PublicKey:         [48]byte{1, 2, 3},
			ProofOfPossession: [96]byte{1, 2, 3},
			WithdrawalAddress: [20]byte{1, 3, 4},
		},
	}

	sigDecode, _ := hex.DecodeString("ae09507041b2ccb9e3b3f9cda71ffae3dc8b2c83f331ebdc98cc4269c56bd4db05706bf317c8877608bc751b36d9af380c5fea6bc804d2080940b3910acc8f222fc4b59166630d8a3b31eba539325c2c60aaaa0408e986241cb462fad8652bdc")
	sigBls, _ := bls.SignatureFromBytes(sigDecode)
	pubDecode, _ := hex.DecodeString("8509d515b24c5a728b26a1b03b023238616dc62d1760f07b90b37407c3535f3fcf4f412dcffa400e4f2b142285c18157")
	pubBls, _ := bls.PublicKeyFromBytes(pubDecode)
	var sig [96]byte
	var pub [48]byte
	copy(sig[:], sigBls.Marshal())
	copy(pub[:], pubBls.Marshal())
	d.PublicKey = pub
	d.Signature = sig

	assert.Equal(t, "424e42e029875876927f3fe7aa8753d8c2a4dd5e7939d0a140d268ad79f30e83", d.Hash().String())

	retSig, err := d.GetSignature()
	assert.NoError(t, err)
	assert.Equal(t, retSig, sigBls)

	retPub, err := d.GetPublicKey()
	assert.NoError(t, err)
	assert.Equal(t, retPub, pubBls)
}

func TestDepositData(t *testing.T) {
	d := testdata.FuzzDepositData()
	ser, err := d.Marshal()
	assert.NoError(t, err)

	desc := new(primitives.DepositData)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, d, desc)

	sigDecode, _ := hex.DecodeString("ae09507041b2ccb9e3b3f9cda71ffae3dc8b2c83f331ebdc98cc4269c56bd4db05706bf317c8877608bc751b36d9af380c5fea6bc804d2080940b3910acc8f222fc4b59166630d8a3b31eba539325c2c60aaaa0408e986241cb462fad8652bdc")
	sigBls, _ := bls.SignatureFromBytes(sigDecode)
	pubDecode, _ := hex.DecodeString("8509d515b24c5a728b26a1b03b023238616dc62d1760f07b90b37407c3535f3fcf4f412dcffa400e4f2b142285c18157")
	pubBls, _ := bls.PublicKeyFromBytes(pubDecode)
	var sig [96]byte
	var pub [48]byte
	copy(sig[:], sigBls.Marshal())
	copy(pub[:], pubBls.Marshal())
	d.ProofOfPossession = sig
	d.PublicKey = pub

	retSig, err := d.GetSignature()
	assert.NoError(t, err)
	assert.Equal(t, retSig, sigBls)

	retPub, err := d.GetPublicKey()
	assert.NoError(t, err)
	assert.Equal(t, retPub, pubBls)
}
