// Copyright (c) 2019 Phore Project

package bls_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

func TestBasicSignature(t *testing.T) {
	s := bls.RandKey()

	p := s.PublicKey()

	msg := []byte("test!")

	sig := s.Sign(msg)

	valid := sig.Verify(msg, p)

	assert.True(t, valid)
}

type XORShift struct {
	state uint64
}

func NewXORShift(state uint64) *XORShift {
	return &XORShift{state}
}

func (xor *XORShift) Read(b []byte) (int, error) {
	for i := range b {
		x := xor.state
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
		b[i] = uint8(x)
		xor.state = x
	}
	return len(b), nil
}

func TestAggregateSignatures(t *testing.T) {
	s0 := bls.RandKey()
	s1 := bls.RandKey()
	s2 := bls.RandKey()

	p0 := s0.PublicKey()
	p1 := s1.PublicKey()
	p2 := s2.PublicKey()

	msg := chainhash.HashH([]byte("test!"))

	sig0 := s0.Sign(msg[:])
	sig1 := s1.Sign(msg[:])
	sig2 := s2.Sign(msg[:])

	aggregateSig := bls.AggregateSignatures([]*bls.Signature{sig0, sig1, sig2})

	valid := aggregateSig.FastAggregateVerify([]*bls.PublicKey{p0, p1, p2}, msg)

	assert.True(t, valid)

}

// func TestVerifyAggregate(t *testing.T) {
// 	r := NewXORShift(1)

// 	s0, _ := bls.RandSecretKey(r)
// 	s1, _ := bls.RandSecretKey(r)
// 	s2, _ := bls.RandSecretKey(r)

// 	p0 := s0.DerivePublicKey()
// 	p1 := s1.DerivePublicKey()
// 	p2 := s2.DerivePublicKey()

// 	msg0 := []byte("test!")
// 	msg1 := []byte("test! 1")
// 	msg2 := []byte("test! 2")

// 	sig0, err := bls.Sign(s0, msg0)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	sig1, err := bls.Sign(s1, msg1)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	sig2, err := bls.Sign(s2, msg2)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	aggregateSig, err := bls.AggregateSigs([]*bls.Signature{sig0, sig1, sig2})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	valid := bls.VerifyAggregate([]*bls.PublicKey{p0, p1, p2}, [][]byte{msg0, msg1, msg2}, aggregateSig)
// 	if !valid {
// 		t.Fatal("aggregate signature was not valid")
// 	}
// }

// func TestVerifyAggregateSeparate(t *testing.T) {
// 	r := NewXORShift(1)

// 	s0, _ := bls.RandSecretKey(r)
// 	s1, _ := bls.RandSecretKey(r)
// 	s2, _ := bls.RandSecretKey(r)

// 	p0 := s0.DerivePublicKey()
// 	p1 := s1.DerivePublicKey()
// 	p2 := s2.DerivePublicKey()

// 	msg0 := []byte("test!")

// 	sig0, err := bls.Sign(s0, msg0)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	sig1, err := bls.Sign(s1, msg0)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	sig2, err := bls.Sign(s2, msg0)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	aggregateSig, err := bls.AggregateSigs([]*bls.Signature{sig0, sig1, sig2})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	aggPk := bls.NewAggregatePublicKey()
// 	aggPk.AggregatePubKey(p0)
// 	aggPk.AggregatePubKey(p1)
// 	aggPk.AggregatePubKey(p2)

// 	valid, err := bls.VerifySig(aggPk, msg0, aggregateSig)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !valid {
// 		t.Fatal("aggregate signature was not valid")
// 	}

// 	aggPk = bls.AggregatePubKeys([]*bls.PublicKey{p0, p1, p2})
// 	valid, err = bls.VerifySig(aggPk, msg0, aggregateSig)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !valid {
// 		t.Fatal("aggregate signature was not valid")
// 	}

// 	aggregateSig = bls.NewAggregateSignature()
// 	aggregateSig.AggregateSig(sig0)
// 	aggregateSig.AggregateSig(sig1)
// 	aggregateSig.AggregateSig(sig2)
// 	valid, err = bls.VerifySig(aggPk, msg0, aggregateSig)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !valid {
// 		t.Fatal("aggregate signature was not valid")
// 	}
// }

// func TestSerializeDeserializeSignature(t *testing.T) {
// 	r := NewXORShift(1)

// 	k, _ := bls.RandSecretKey(r)
// 	pub := k.DerivePublicKey()

// 	sig, err := bls.Sign(k, []byte("testing!"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	sigAfter, err := bls.DeserializeSignature(sig.Serialize())
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	valid, err := bls.VerifySig(pub, []byte("testing!"), sigAfter)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if !valid {
// 		t.Fatal("signature did not verify")
// 	}
// }

// func TestSerializeDeserializeSecret(t *testing.T) {
// 	r := NewXORShift(1)

// 	k, _ := bls.RandSecretKey(r)
// 	pub := k.DerivePublicKey()

// 	kSer := k.Serialize()
// 	kNew := bls.DeserializeSecretKey(kSer)

// 	sig, err := bls.Sign(&kNew, chainhash.HashB([]byte("testing!")))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	valid, err := bls.VerifySig(pub, chainhash.HashB([]byte("testing!")), sig)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if !valid {
// 		t.Fatal("signature did not verify")
// 	}
// }

// func TestCopyPubkey(t *testing.T) {
// 	r := NewXORShift(1)

// 	k, _ := bls.RandSecretKey(r)

// 	p := k.DerivePublicKey()
// 	p2 := p.Copy()

// 	p.Copy()

// 	p.AggregatePubKey(&p2)

// 	if p2.Equals(*p) {
// 		t.Fatal("pubkey copy is incorrect")
// 	}
// }

// func TestCopySignature(t *testing.T) {
// 	r := NewXORShift(1)

// 	k, _ := bls.RandSecretKey(r)

// 	s, err := bls.Sign(k, []byte{})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	s2, err := bls.Sign(k, []byte{})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	sCopy := s.Copy()

// 	s.AggregateSig(s2)
// 	sSer := s.Serialize()
// 	sCopySer := sCopy.Serialize()
// 	if bytes.Equal(sSer[:], sCopySer[:]) {
// 		t.Fatal("copy returns pointer")
// 	}
// }

// func TestNewSecretFromBech32(t *testing.T) {
// 	r := NewXORShift(1)
// 	k, _ := bls.RandSecretKey(r)
// 	encSecret := k.ToBech32(params.Mainnet.AddressPrefixes, false)
// 	secret, err := bls.NewSecretFromBech32(encSecret, params.Mainnet.AddressPrefixes, false)
// 	if err != nil {
// 		t.Fatal("unable to get secret from bech32")
// 	}
// 	equal := reflect.DeepEqual(k, &secret)
// 	if !equal {
// 		t.Fatal("keys doesn't match")
// 	}
// 	encSecretContract := k.ToBech32(params.Mainnet.AddressPrefixes, true)
// 	secretContract, err := bls.NewSecretFromBech32(encSecretContract, params.Mainnet.AddressPrefixes, true)
// 	if err != nil {
// 		t.Fatal("unable to get secret from bech32")
// 	}
// 	equal = reflect.DeepEqual(k, &secretContract)
// 	if !equal {
// 		t.Fatal("keys doesn't match")
// 	}
// }

// func TestPubKeyHash(t *testing.T) {
// 	pubKeyBytes, err := hex.DecodeString("b77abbbf316558e0b4c3d1aa9e0692a25d19e856e3f763dcb6476f7f5fe50d82a69227eaef718aa66b13bfe131388e8e")
// 	if err != nil {
// 		t.Fatal("unable to get decode public string")
// 	}
// 	var serPubKey [48]byte
// 	buf := bytes.NewBuffer(serPubKey[:0])
// 	buf.Write(pubKeyBytes)
// 	pubKey, err := bls.DeserializePublicKey(serPubKey)
// 	if err != nil {
// 		t.Fatal("unable to get deserialize public key")
// 	}
// 	encPub, err := pubKey.ToBech32(params.Mainnet.AddressPrefixes, false)
// 	if err != nil {
// 		t.Fatal("unable to get public bech32")
// 	}
// 	equal := reflect.DeepEqual(encPub, "olpub1yne583dks9ptymxya4dkakkx0sd2kyz58umv42wrt9vfq3xlkqgsh0xff2")
// 	if !equal {
// 		t.Fatal("pubKeyHashes doesn't match")
// 	}
// 	encPubContract, err := pubKey.ToBech32(params.Mainnet.AddressPrefixes, true)
// 	if err != nil {
// 		t.Fatal("unable to get public bech32")
// 	}
// 	equal = reflect.DeepEqual(encPubContract, "ctpub1rdden82dqeks8ajkgxajwfaxe52zdjmpxkvc3decrl98kk8xghcqq3dz88")
// 	if !equal {
// 		t.Fatal("pubKeyHashes doesn't match")
// 	}
// }
