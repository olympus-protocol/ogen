package burnproof_test

import (
	"bytes"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

var merkleRootHashTestNet [32]byte

func init() {
	hashBytes, _ := hex.DecodeString("0be71bd3e3ec9046901b21f066407f8413c73dc3145d6a515d0fa03b28e0140f") //  PolisBlockchain "height": 750711
	copy(merkleRootHashTestNet[:], hashBytes)
}

var proofBytes, _ = ioutil.ReadFile("./test_proofdata.dat")

func TestCoinProofDecode(t *testing.T) {
	buf := bytes.NewBuffer(proofBytes)

	var proofs []*burnproof.CoinsProof
	for {
		coinProof := new(burnproof.CoinsProof)
		err := coinProof.Unmarshal(buf)
		assert.NoError(t, err)
		proofs = append(proofs, coinProof)
		if buf.Len() <= 0 {
			break
		}
	}
}

func TestBurnVerify(t *testing.T) {
	acc := []byte("12345")

	err := burnproof.VerifyBurn(proofBytes, acc, merkleRootHashTestNet)
	assert.NoError(t, err)
}

func TestBurnProofsToSerializable(t *testing.T) {
	buf := bytes.NewBuffer(proofBytes)

	var proofs []*burnproof.CoinsProof
	for {
		coinProof := new(burnproof.CoinsProof)
		err := coinProof.Unmarshal(buf)
		assert.NoError(t, err)
		proofs = append(proofs, coinProof)
		if buf.Len() <= 0 {
			break
		}
	}
	acc := []byte("12345")
	var accBytes [44]byte
	copy(accBytes[:], acc)

	amount := uint64(0)
	for _, proof := range proofs {
		serProof, err := proof.ToSerializable(accBytes)
		assert.NoError(t, err)

		ser, err := serProof.Marshal()
		assert.NoError(t, err)
		var newSerProof burnproof.CoinsProofSerializable

		err = newSerProof.Unmarshal(ser)
		assert.NoError(t, err)

		toCoinProof, err := newSerProof.ToCoinProof()
		assert.NoError(t, err)

		amount += uint64(toCoinProof.Transaction.TxOut[0].Value)
		err = burnproof.VerifyBurn(proofBytes, acc, merkleRootHashTestNet)
		assert.NoError(t, err)

		assert.Equal(t, proof, toCoinProof)
	}
}
