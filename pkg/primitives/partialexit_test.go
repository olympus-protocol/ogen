package primitives_test

import (
	"encoding/hex"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPartialExit(t *testing.T) {
	v := testdata.FuzzPartialExits(10)
	for _, e := range v {
		ser, err := e.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.PartialExit)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, e, desc)
	}

	e := new(primitives.PartialExit)

	sigDecode, _ := hex.DecodeString("ae09507041b2ccb9e3b3f9cda71ffae3dc8b2c83f331ebdc98cc4269c56bd4db05706bf317c8877608bc751b36d9af380c5fea6bc804d2080940b3910acc8f222fc4b59166630d8a3b31eba539325c2c60aaaa0408e986241cb462fad8652bdc")
	sigBls, _ := bls.SignatureFromBytes(sigDecode)
	pubDecode, _ := hex.DecodeString("8509d515b24c5a728b26a1b03b023238616dc62d1760f07b90b37407c3535f3fcf4f412dcffa400e4f2b142285c18157")
	pubBls, _ := bls.PublicKeyFromBytes(pubDecode)
	var sig [96]byte
	var pub [48]byte
	copy(sig[:], sigBls.Marshal())
	copy(pub[:], pubBls.Marshal())
	e.Signature = sig
	e.WithdrawPubkey = pub
	e.ValidatorPubkey = pub

	getValPub, _ := e.GetValidatorPubKey()
	getWithPub, _ := e.GetWithdrawPubKey()
	getSig, _ := e.GetSignature()

	assert.Equal(t, pubBls, getValPub)
	assert.Equal(t, pubBls, getWithPub)
	assert.Equal(t, sigBls, getSig)

	assert.Equal(t, "d93da4cb8b746301ab77e11d0fb2445ee93589ac507bc1856fdebb36c1f51762", e.Hash().String())
}
