package bls_test

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CombinedSignatureCopy(t *testing.T) {
	rand := bls.RandKey()
	comb := bls.NewCombinedSignature(rand.PublicKey(), rand.Sign([]byte("test")))
	comb2 := comb.Copy()

	newKey := bls.RandKey()

	var newPub [48]byte
	copy(newPub[:], newKey.PublicKey().Marshal())
	var newSig [96]byte
	copy(newSig[:], newKey.Sign([]byte("test")).Marshal())
	comb.P = newPub
	comb.S = newSig

	pub, err := comb2.GetPublicKey()
	assert.NoError(t, err)

	var copyNewPub [48]byte
	copy(copyNewPub[:], pub.Marshal())

	assert.Equal(t, copyNewPub[:], rand.PublicKey().Marshal())
}
