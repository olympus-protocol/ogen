package bls_kilic_test

import (
	"bytes"
	"github.com/olympus-protocol/ogen/pkg/bls/bls_kilic"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

var impl bls_kilic.KilicImplementation

func TestBasicSignature(t *testing.T) {

	s := impl.RandKey()

	p := s.PublicKey()

	msg := []byte("test!")

	sig := s.Sign(msg)

	valid := sig.Verify(p, msg)

	assert.True(t, valid)
}

func TestAggregateSignatures(t *testing.T) {
	//s0 := impl.RandKey()
	//s1 := impl.RandKey()
	//s2 := impl.RandKey()

	//p0 := s0.PublicKey()
	//p1 := s1.PublicKey()
	//p2 := s2.PublicKey()

	//msg := chainhash.HashH([]byte("test!"))

	//sig0 := s0.Sign(msg[:])
	//sig1 := s1.Sign(msg[:])
	//sig2 := s2.Sign(msg[:])

	//aggregateSig := impl.AggregateSignatures([]bls_interface.Signature{sig0, sig1, sig2})

	//valid := aggregateSig.FastAggregateVerify([]bls_interface.PublicKey{p0, p1, p2}, msg)

	//assert.True(t, valid)

}

func TestVerifyAggregate(t *testing.T) {

	//s0 := impl.RandKey()
	//s1 := impl.RandKey()
	//s2 := impl.RandKey()

	//p0 := s0.PublicKey()
	//p1 := s1.PublicKey()
	//p2 := s2.PublicKey()

	//msg0 := [32]byte{0x1}
	//msg1 := [32]byte{0x2}
	//msg2 := [32]byte{0x3}

	//sig0 := s0.Sign(msg0[:])

	//sig1 := s1.Sign(msg1[:])

	//sig2 := s2.Sign(msg2[:])

	//var sigs [][]byte
	//sigs = append(sigs, sig0.Marshal())
	//sigs = append(sigs, sig1.Marshal())
	//sigs = append(sigs, sig2.Marshal())

	//valid, err := impl.VerifyMultipleSignatures(sigs, [][32]byte{msg0, msg1, msg2}, []bls_interface.PublicKey{p0, p1, p2})
	//assert.NoError(t, err)
	//assert.True(t, valid)
}

func TestSerializeDeserializeSignature(t *testing.T) {
	impl := bls_kilic.KilicImplementation{}

	k := impl.RandKey()
	p := k.PublicKey()

	sig := k.Sign([]byte("testing!"))

	sigAfter, err := impl.SignatureFromBytes(sig.Marshal())
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, sigAfter.Verify(p, []byte("testing!")))

}

func TestSerializeDeserializeSecret(t *testing.T) {

	k := impl.RandKey()
	pub := k.PublicKey()

	kSer := k.Marshal()

	kNew, err := impl.SecretKeyFromBytes(kSer)
	assert.NoError(t, err)

	sig := kNew.Sign(chainhash.HashB([]byte("testing!")))

	assert.True(t, sig.Verify(pub, chainhash.HashB([]byte("testing!"))))
}

func TestCopyPubkey(t *testing.T) {
	k := impl.RandKey()

	p := k.PublicKey()
	p2 := p.Copy()

	p.Copy()

	p.Aggregate(p2)

	assert.False(t, bytes.Equal(p.Marshal(), p2.Marshal()))
}

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
