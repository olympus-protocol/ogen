package primitives_test

import (
	"encoding/hex"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVoteSlashing(t *testing.T) {
	v := testdata.FuzzVoteSlashing(10, true, true)
	for _, c := range v {
		ser, err := c.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.VoteSlashing)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)
	}

	incorrect := testdata.FuzzVoteSlashing(10, false, true)

	for _, c := range incorrect {
		_, err := c.Marshal()
		assert.NotNil(t, err)
	}

	nildata := testdata.FuzzVoteSlashing(10, true, false)

	for _, c := range nildata {
		assert.NotPanics(t, func() {
			data, err := c.Marshal()
			assert.NoError(t, err)

			n := new(primitives.VoteSlashing)
			err = n.Unmarshal(data)
			assert.NoError(t, err)

			assert.Equal(t, c, n)

			assert.Equal(t, uint64(0), n.Vote1.Data.Slot)
			assert.Equal(t, uint64(0), n.Vote2.Data.Slot)

		})
	}

	d := primitives.VoteSlashing{
		Vote1: &primitives.MultiValidatorVote{
			Data: &primitives.VoteData{
				Slot:            5,
				FromEpoch:       5,
				FromHash:        [32]byte{1, 2, 3},
				ToEpoch:         5,
				ToHash:          [32]byte{1, 2, 3},
				BeaconBlockHash: [32]byte{1, 2, 3},
				Nonce:           5,
			},
			Sig:                   [96]byte{1, 2, 3},
			ParticipationBitfield: bitfield.NewBitlist(6042),
		},
		Vote2: &primitives.MultiValidatorVote{
			Data: &primitives.VoteData{
				Slot:            5,
				FromEpoch:       5,
				FromHash:        [32]byte{1, 2, 3},
				ToEpoch:         5,
				ToHash:          [32]byte{1, 2, 3},
				BeaconBlockHash: [32]byte{1, 2, 3},
				Nonce:           5,
			},
			Sig:                   [96]byte{1, 2, 3},
			ParticipationBitfield: bitfield.NewBitlist(6042),
		},
	}

	assert.Equal(t, "0299f30f1dab1026bfc3f1179631fa3af0eb9f0ee0b52ee423d344203c02522a", d.Hash().String())

}

func TestRANDAOSlashing(t *testing.T) {
	v := testdata.FuzzRANDAOSlashing(10)
	for _, c := range v {
		ser, err := c.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.RANDAOSlashing)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)
	}

	d := primitives.RANDAOSlashing{
		RandaoReveal:    [96]byte{},
		Slot:            100,
		ValidatorPubkey: [48]byte{},
	}

	sigDecode, _ := hex.DecodeString("ae09507041b2ccb9e3b3f9cda71ffae3dc8b2c83f331ebdc98cc4269c56bd4db05706bf317c8877608bc751b36d9af380c5fea6bc804d2080940b3910acc8f222fc4b59166630d8a3b31eba539325c2c60aaaa0408e986241cb462fad8652bdc")
	sigBls, _ := bls.SignatureFromBytes(sigDecode)
	pubDecode, _ := hex.DecodeString("8509d515b24c5a728b26a1b03b023238616dc62d1760f07b90b37407c3535f3fcf4f412dcffa400e4f2b142285c18157")
	pubBls, _ := bls.PublicKeyFromBytes(pubDecode)
	var sig [96]byte
	var pub [48]byte
	copy(sig[:], sigBls.Marshal())
	copy(pub[:], pubBls.Marshal())
	d.RandaoReveal = sig
	d.ValidatorPubkey = pub

	assert.Equal(t, "d3b0e8858d4add359a53a8604808de3abaa0d0f13d64e6017afae8cbbca32414", d.Hash().String())

	retSig, err := d.GetRandaoReveal()
	assert.NoError(t, err)
	assert.Equal(t, sigBls, retSig)

	retPub, err := d.GetValidatorPubkey()
	assert.NoError(t, err)
	assert.Equal(t, pubBls, retPub)

}

func TestProposerSlashing(t *testing.T) {
	v := testdata.FuzzProposerSlashing(10, true)
	for _, c := range v {
		ser, err := c.Marshal()
		assert.NoError(t, err)

		desc := new(primitives.ProposerSlashing)
		err = desc.Unmarshal(ser)
		assert.NoError(t, err)

		assert.Equal(t, c, desc)
	}

	nildata := testdata.FuzzProposerSlashing(10, false)

	for _, c := range nildata {
		assert.NotPanics(t, func() {
			data, err := c.Marshal()
			assert.NoError(t, err)

			n := new(primitives.ProposerSlashing)
			err = n.Unmarshal(data)
			assert.NoError(t, err)

			assert.Equal(t, c, n)

			assert.Equal(t, uint64(0), n.BlockHeader1.Slot)
			assert.Equal(t, uint64(0), n.BlockHeader2.Slot)

		})
	}

	d := primitives.ProposerSlashing{
		BlockHeader1: &primitives.BlockHeader{
			Version:                    0,
			Nonce:                      0,
			TxMerkleRoot:               [32]byte{},
			VoteMerkleRoot:             [32]byte{},
			DepositMerkleRoot:          [32]byte{},
			ExitMerkleRoot:             [32]byte{},
			VoteSlashingMerkleRoot:     [32]byte{},
			RANDAOSlashingMerkleRoot:   [32]byte{},
			ProposerSlashingMerkleRoot: [32]byte{},
			GovernanceVotesMerkleRoot:  [32]byte{},
			PrevBlockHash:              [32]byte{},
			Timestamp:                  0,
			Slot:                       0,
			StateRoot:                  [32]byte{},
			FeeAddress:                 [20]byte{},
		},
		BlockHeader2: &primitives.BlockHeader{
			Version:                    0,
			Nonce:                      0,
			TxMerkleRoot:               [32]byte{},
			VoteMerkleRoot:             [32]byte{},
			DepositMerkleRoot:          [32]byte{},
			ExitMerkleRoot:             [32]byte{},
			VoteSlashingMerkleRoot:     [32]byte{},
			RANDAOSlashingMerkleRoot:   [32]byte{},
			ProposerSlashingMerkleRoot: [32]byte{},
			GovernanceVotesMerkleRoot:  [32]byte{},
			PrevBlockHash:              [32]byte{},
			Timestamp:                  0,
			Slot:                       0,
			StateRoot:                  [32]byte{},
			FeeAddress:                 [20]byte{},
		},
	}

	assert.Equal(t, "de2ca1d3d1c0862b06e5e5e33aefe44a2d0290e4ef6228128a1e797e35d3457c", d.Hash().String())

	sigDecode, _ := hex.DecodeString("ae09507041b2ccb9e3b3f9cda71ffae3dc8b2c83f331ebdc98cc4269c56bd4db05706bf317c8877608bc751b36d9af380c5fea6bc804d2080940b3910acc8f222fc4b59166630d8a3b31eba539325c2c60aaaa0408e986241cb462fad8652bdc")
	sigBls, _ := bls.SignatureFromBytes(sigDecode)
	pubDecode, _ := hex.DecodeString("8509d515b24c5a728b26a1b03b023238616dc62d1760f07b90b37407c3535f3fcf4f412dcffa400e4f2b142285c18157")
	pubBls, _ := bls.PublicKeyFromBytes(pubDecode)
	var sig [96]byte
	var pub [48]byte
	copy(sig[:], sigBls.Marshal())
	copy(pub[:], pubBls.Marshal())
	d.Signature1 = sig
	d.Signature2 = sig
	d.ValidatorPublicKey = pub

	retSig1, err := d.GetSignature1()
	assert.NoError(t, err)
	assert.Equal(t, retSig1, sigBls)

	retSig2, err := d.GetSignature2()
	assert.NoError(t, err)
	assert.Equal(t, retSig2, sigBls)

	retPub, err := d.GetValidatorPubkey()
	assert.NoError(t, err)
	assert.Equal(t, retPub, pubBls)

}
