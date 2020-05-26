package bls_test

import (
	"testing"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
)

func TestCorrectnessMultisig(t *testing.T) {
	secretKeys := make([]*bls.SecretKey, 20)
	publicKeys := make([]*bls.PublicKey, 20)

	for i := range secretKeys {
		secretKeys[i] = bls.RandKey()
		publicKeys[i] = secretKeys[i].PublicKey()
	}

	// create 10-of-20 multipub
	multiPub := bls.NewMultipub(publicKeys, 10)
	multisig := bls.NewMultisig(multiPub, []byte("hello there!"))

	for i := 0; i < 9; i++ {
		if err := multisig.Sign(secretKeys[i]); err != nil {
			t.Fatal(err)
		}
	}

	if multisig.Verify() {
		t.Fatal("multisig should not validate with less than num needed")
	}

	if err := multisig.Sign(secretKeys[9]); err != nil {
		t.Fatal(err)
	}

	if !multisig.Verify() {
		t.Fatal("multisig should validate with equal to num needed")
	}

	for i := 10; i < 20; i++ {
		if err := multisig.Sign(secretKeys[i]); err != nil {
			t.Fatal(err)
		}
	}

	if !multisig.Verify() {
		t.Fatal("multisig should validate with all pubkeys")
	}

	multiPub.ToBech32(params.AddrPrefixes{
		Multisig: "olmul",
	})
}
