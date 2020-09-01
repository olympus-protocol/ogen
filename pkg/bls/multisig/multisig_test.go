package multisig_test

import (
	"encoding/hex"
	"github.com/olympus-protocol/ogen/pkg/bls/multisig"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/pkg/bls"
)

func init() {
	bls.Initialize(testdata.TestParams)
}

func TestCorrectnessMultisig(t *testing.T) {
	secretKeys := make([]*bls.SecretKey, 20)
	publicKeys := make([]*bls.PublicKey, 20)
	var err error
	for i := range secretKeys {
		secretKeys[i] = bls.RandKey()
		publicKeys[i] = secretKeys[i].PublicKey()
	}

	// create 10-of-20 multipub
	multiPub := multisig.NewMultipub(publicKeys, 10)

	ser, err := multiPub.Marshal()
	assert.NoError(t, err)

	desc := new(multisig.Multipub)

	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	ms := multisig.NewMultisig(multiPub)

	msg := []byte("hello there!")

	for i := 0; i < 9; i++ {
		assert.NoError(t, ms.Sign(secretKeys[i], msg))
	}

	assert.False(t, ms.Verify(msg))

	assert.NoError(t, ms.Sign(secretKeys[9], msg))

	assert.True(t, ms.Verify(msg))

	for i := 10; i < 20; i++ {
		assert.NoError(t, ms.Sign(secretKeys[i], msg))
	}

	assert.True(t, ms.Verify(msg))

	pub, err := ms.GetPublicKey()
	assert.NoError(t, err)


	assert.Equal(t, desc, pub)
}

func TestMultisigSerializeSign(t *testing.T) {
	secretKeys := make([]*bls.SecretKey, 20)
	publicKeys := make([]*bls.PublicKey, 20)

	var err error

	for i := range secretKeys {
		secretKeys[i] = bls.RandKey()
		publicKeys[i] = secretKeys[i].PublicKey()
	}

	// create 10-of-20 multipub
	multiPub := multisig.NewMultipub(publicKeys, 10)
	ms := multisig.NewMultisig(multiPub)

	msg := []byte("hello there!")

	for i := 0; i < 10; i++ {
		assert.NoError(t, ms.Sign(secretKeys[i], msg))
	}

	multiBytes, err := ms.Marshal()

	assert.NoError(t, err)

	newMulti := new(multisig.Multisig)

	assert.NoError(t, newMulti.Unmarshal(multiBytes))

	assert.True(t, newMulti.Verify(msg))
}

func TestMultipubCopy(t *testing.T) {
	secretKeys := make([]*bls.SecretKey, 20)
	publicKeys := make([]*bls.PublicKey, 20)
	for i := range secretKeys {
		secretKeys[i] = bls.RandKey()
		publicKeys[i] = secretKeys[i].PublicKey()
	}

	mp := multisig.NewMultipub(publicKeys, 10)

	cp := mp.Copy()

	mp.NumNeeded = 1
	assert.Equal(t, uint64(10), cp.NumNeeded)

	mp.PublicKeys = nil
	assert.Equal(t, 20, len(cp.PublicKeys))
}

func TestMultisigCopy(t *testing.T) {
	secretKeys := make([]*bls.SecretKey, 20)
	publicKeys := make([]*bls.PublicKey, 20)
	for i := range secretKeys {
		secretKeys[i] = bls.RandKey()
		publicKeys[i] = secretKeys[i].PublicKey()
	}

	mp := multisig.NewMultipub(publicKeys, 10)
	ms := multisig.NewMultisig(mp)

	msg := []byte("hello there!")

	for i := 0; i < 10; i++ {
		assert.NoError(t, ms.Sign(secretKeys[i], msg))
	}

	cp := ms.Copy()

	ms.KeysSigned.Set(11)
	assert.False(t, cp.KeysSigned.Get(11))

	ms.Signatures = nil
	assert.Equal(t, 10, len(cp.Signatures))

	ms.PublicKey = nil
	assert.NotNil(t, cp.PublicKey)
}

func TestMultipubHashing(t *testing.T) {
	var pubs []*bls.PublicKey

	pub1, err := hex.DecodeString("97d2b427a0914f325aa1502064fc724c71400a035c264a20595f4e234fd68a6494a342e7a631a97d63ffcd795763cfe8")
	assert.NoError(t, err)
	pubBls1, err := bls.PublicKeyFromBytes(pub1)
	assert.NoError(t, err)
	pubs = append(pubs, pubBls1)

	pub2, err := hex.DecodeString("a63d0304a1791d636efb22992deae5e23a87267a3a14c77bb8611f60514c916a21d990308840d2e1f048805761165cb8")
	assert.NoError(t, err)
	pubBls2, err := bls.PublicKeyFromBytes(pub2)
	assert.NoError(t, err)
	pubs = append(pubs, pubBls2)

	pub3, err := hex.DecodeString("b6af7838a151583d9a424acdd333a7f97c9d394a63b68fa0d2e7dcd82dc8af97a7115590ca121cf6863c49c86f8ce8b8")
	assert.NoError(t, err)
	pubBls3, err := bls.PublicKeyFromBytes(pub3)
	assert.NoError(t, err)
	pubs = append(pubs, pubBls3)

	pub4, err := hex.DecodeString("b7ce6ce7743bdbde6d264f7c5ecc642cdbbe1208588229f742d529bc2a407e962c98810a18795daced7fcb0a0144e03a")
	assert.NoError(t, err)
	pubBls4, err := bls.PublicKeyFromBytes(pub4)
	assert.NoError(t, err)
	pubs = append(pubs, pubBls4)

	pub5, err := hex.DecodeString("8f4f4d7aa6fa2297d1cf129539a212521484c3ad59b6e001354ea31dc8a30c9f31e29c72e152a64c7d87515c87473eac")
	assert.NoError(t, err)
	pubBls5, err := bls.PublicKeyFromBytes(pub5)
	assert.NoError(t, err)
	pubs = append(pubs, pubBls5)

	pub6, err := hex.DecodeString("8839e6b20e5de6d696fb7b4bc6397c37ef82002caf7eba99889a10c2be3a6aa73a25486d399f84d0d774505252bbb156")
	assert.NoError(t, err)
	pubBls6, err := bls.PublicKeyFromBytes(pub6)
	assert.NoError(t, err)
	pubs = append(pubs, pubBls6)

	pub7, err := hex.DecodeString("afc7d556ced1abbfbc09993dbc949e7af64a067e8df66215a1bd0a0bca1e2b51b811f6002c82e17e67063c547a2bffae")
	assert.NoError(t, err)
	pubBls7, err := bls.PublicKeyFromBytes(pub7)
	assert.NoError(t, err)
	pubs = append(pubs, pubBls7)

	pub8, err := hex.DecodeString("849842e33f40c13b50c52a93e326d059ee171d90bf19167e545ae1ff2021e351b003d0b86b7027447a0f9610f384cec6")
	assert.NoError(t, err)
	pubBls8, err := bls.PublicKeyFromBytes(pub8)
	assert.NoError(t, err)
	pubs = append(pubs, pubBls8)

	pub9, err := hex.DecodeString("ae21bea1828c901e750ab57c177d9fbe0ca7cbc9170c2efacf1aadf609d2a80abb56f4684e45fd28a0af5aa6993205a5")
	assert.NoError(t, err)
	pubBls9, err := bls.PublicKeyFromBytes(pub9)
	assert.NoError(t, err)
	pubs = append(pubs, pubBls9)

	pub10, err := hex.DecodeString("aac185face768f8a7a87af886b17c62ff579e85eae39a7a4dbebc01e91ef26045048d8343dd38412a6070a73ac7788a7")
	assert.NoError(t, err)
	pubBls10, err := bls.PublicKeyFromBytes(pub10)
	assert.NoError(t, err)
	pubs = append(pubs, pubBls10)

	mp := multisig.NewMultipub(pubs, 10)

	hash, err := mp.Hash()
	assert.NoError(t, err)

	assert.Equal(t, "2702d01577ca30238e4a2e261496f757e11597d3", hex.EncodeToString(hash[:]))

	acc, err := mp.ToBech32()
	assert.Equal(t, "itpub1yupdq9thegcz8rj29cnpf9hh2ls3t97np4smcp", acc)

}