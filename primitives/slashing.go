package primitives

import (
	"bytes"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// VoteSlashing is a slashing where validators vote in the span of their
// other votes.
type VoteSlashing struct {
	Vote1 MultiValidatorVote
	Vote2 MultiValidatorVote
}

// Encode encodes the vote slashing to a writer.
func (vs *VoteSlashing) Encode(w io.Writer) error {
	if err := vs.Vote1.Serialize(w); err != nil {
		return err
	}
	if err := vs.Vote2.Serialize(w); err != nil {
		return err
	}

	return nil
}

// Decode decodes the vote slashing to the given reader.
func (vs *VoteSlashing) Decode(r io.Reader) error {
	if err := vs.Vote1.Deserialize(r); err != nil {
		return err
	}
	return vs.Vote2.Deserialize(r)
}

// Hash calculates the hash of the slashing.
func (vs *VoteSlashing) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = vs.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}

// RANDAOSlashing is a slashing where a validator reveals their RANDAO
// signature too early.
type RANDAOSlashing struct {
	RandaoReveal    bls.Signature
	Slot            uint64
	ValidatorPubkey bls.PublicKey
}

// Encode encodes the RANDAOSlashing to the given writer.
func (rs *RANDAOSlashing) Encode(w io.Writer) error {
	sigBytes := rs.RandaoReveal.Marshal()
	pubkeyBytes := rs.ValidatorPubkey.Marshal()

	return serializer.WriteElements(w, sigBytes, pubkeyBytes, rs.Slot)
}

// Decode decodes the RANDAOSlashing from the given reader.
func (rs *RANDAOSlashing) Decode(r io.Reader) error {
	sigBytes := make([]byte, 96)
	pubBytes := make([]byte, 48)

	if err := serializer.ReadElements(r, &sigBytes, &pubBytes, &rs.Slot); err != nil {
		return err
	}

	sig, err := bls.SignatureFromBytes(sigBytes)
	if err != nil {
		return err
	}

	pub, err := bls.PublicKeyFromBytes(pubBytes)
	if err != nil {
		return err
	}

	rs.RandaoReveal = *sig
	rs.ValidatorPubkey = *pub

	return nil
}

// Hash calculates the hash of the RANDAO slashing.
func (rs *RANDAOSlashing) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = rs.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}

// ProposerSlashing is a slashing to a block proposer that proposed
// two blocks at the same slot.
type ProposerSlashing struct {
	BlockHeader1       BlockHeader
	BlockHeader2       BlockHeader
	Signature1         bls.Signature
	Signature2         bls.Signature
	ValidatorPublicKey bls.PublicKey
}

// Encode encodes the proposer slashing to the given writer.
func (ps *ProposerSlashing) Encode(w io.Writer) error {
	if err := ps.BlockHeader1.Serialize(w); err != nil {
		return err
	}
	if err := ps.BlockHeader2.Serialize(w); err != nil {
		return err
	}
	if err := ps.Signature1.Encode(w); err != nil {
		return err
	}
	if err := ps.Signature2.Encode(w); err != nil {
		return err
	}
	return ps.ValidatorPublicKey.Encode(w)
}

// Decode decodes the proposer slashing from the given reader.
func (ps *ProposerSlashing) Decode(r io.Reader) error {
	if err := ps.BlockHeader1.Deserialize(r); err != nil {
		return err
	}
	if err := ps.BlockHeader2.Deserialize(r); err != nil {
		return err
	}
	if err := ps.Signature1.Decode(r); err != nil {
		return err
	}
	if err := ps.Signature2.Decode(r); err != nil {
		return err
	}
	return ps.ValidatorPublicKey.Decode(r)
}

// Hash calculates the hash of the proposer slashing.
func (ps *ProposerSlashing) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = ps.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}
