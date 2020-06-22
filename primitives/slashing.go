package primitives

import (
	"bytes"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
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

// Hash calculates the hash of the proposer slashing.
func (ps *ProposerSlashing) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = ps.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}
