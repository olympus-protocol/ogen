package bls_test

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignature_Copy(t *testing.T) {
	rand := bls.RandKey()

	sig := rand.Sign([]byte("test"))

	sig2 := sig.Copy()

	sig = bls.RandKey().Sign([]byte("test"))

	assert.Equal(t, sig2.Marshal(), rand.Sign([]byte("test")).Marshal())
}
