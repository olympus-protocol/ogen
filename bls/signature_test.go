package bls_test

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignature_Copy(t *testing.T) {
	rand := bls.RandKey()

	sig := rand.Sign([]byte("test"))

	sig2 := sig.Copy()

	sig = bls.RandKey().Sign([]byte("test"))

	assert.NotEqual(t, sig2.Marshal(), sig.Marshal())
}