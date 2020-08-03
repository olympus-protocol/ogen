package bls_test

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPublicKey_Copy(t *testing.T) {
	rand := bls.RandKey()

	pub := rand.PublicKey()

	pub2 := pub.Copy()

	pub = bls.RandKey().PublicKey()

	assert.NotEqual(t, pub2.Marshal(), pub.Marshal())
}
