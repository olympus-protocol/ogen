package primitives_test

import (
	"encoding/hex"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTx(t *testing.T) {
	v := testdata.FuzzTx(10)
	for _, c := range v {
		ser, err := c.Marshal()
		assert.NoError(t, err)

		assert.Equal(t, primitives.TxSize, len(ser))

		desc := new(primitives.Tx)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)

		assert.NoError(t, c.VerifySig())
	}

	sigDecode, _ := hex.DecodeString("ae09507041b2ccb9e3b3f9cda71ffae3dc8b2c83f331ebdc98cc4269c56bd4db05706bf317c8877608bc751b36d9af380c5fea6bc804d2080940b3910acc8f222fc4b59166630d8a3b31eba539325c2c60aaaa0408e986241cb462fad8652bdc")
	pubDecode, _ := hex.DecodeString("8509d515b24c5a728b26a1b03b023238616dc62d1760f07b90b37407c3535f3fcf4f412dcffa400e4f2b142285c18157")
	pubBls, _ := bls.PublicKeyFromBytes(pubDecode)

	pubHash, _ := pubBls.Hash()

	tx := primitives.Tx{
		To:     pubHash,
		Amount: 0,
		Nonce:  0,
		Fee:    0,
	}

	copy(tx.Signature[:], sigDecode)
	copy(tx.FromPublicKey[:], pubDecode)

	fromPubKey, _ := tx.FromPubkeyHash()
	assert.Equal(t, pubHash, fromPubKey)

	assert.Equal(t, "9b6b8988800352ae1198cca98ebaaf56f0ba9ff834ed183f4aa8a49596e8e56a", tx.Hash().String())
}
