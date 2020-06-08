package primitives

import (
	"bytes"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// GovernanceVoteType is the type of a governance vote.
type GovernanceVoteType uint8

const (
	// EnterVotingPeriod can be done by anyone on the network to signal that they
	// want a voting period to start.
	EnterVotingPeriod GovernanceVoteType = iota

	// VoteFor can be done by anyone on the network during a voting period to vote
	// for a specific assignment of managers.
	VoteFor

	// UpdateManagersInstantly updates the current managers on the condition that
	// the signature is signed by 5/5 of the managers.
	UpdateManagersInstantly

	// UpdateManagersVote immediately triggers a community vote to re-elect certain
	// managers on the condition that the signature is signed by 3/5 of the managers.
	UpdateManagersVote
)

// GovernanceVote is a vote for governance.
type GovernanceVote struct {
	Type      GovernanceVoteType
	Data      []byte
	Signature bls.FunctionalSignature
}

// Encode encodes the governance vote to a writer.
func (gv *GovernanceVote) Encode(w io.Writer) error {
	if err := serializer.WriteElement(w, gv.Type); err != nil {
		return err
	}
	if err := serializer.WriteVarBytes(w, gv.Data); err != nil {
		return err
	}
	if err := gv.Signature.Encode(w); err != nil {
		return err
	}
	return nil
}

// Decode decodes the governance vote from the reader.
func (gv *GovernanceVote) Decode(r io.Reader) error {
	if err := serializer.ReadElement(r, &gv.Type); err != nil {
		return err
	}
	data, err := serializer.ReadVarBytes(r)
	if err != nil {
		return err
	}
	gv.Data = data
	if err := gv.Signature.Decode(r); err != nil {
		return err
	}

	return nil
}

// Hash calculates the hash of the governance vote.
func (gv *GovernanceVote) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = gv.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}
