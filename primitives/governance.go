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

// CommunityVoteData is the votes that users sign to vote for a specific candidate.
type CommunityVoteData struct {
	ReplacementCandidates [][20]byte
}

// Encode encodes the community vote data to the given writer.
func (c *CommunityVoteData) Encode(w io.Writer) error {
	if err := serializer.WriteVarInt(w, uint64(len(c.ReplacementCandidates))); err != nil {
		return err
	}

	for _, ca := range c.ReplacementCandidates {
		if err := serializer.WriteElement(w, ca); err != nil {
			return err
		}
	}

	return nil
}

// Decode decodes the community vote data from the given reader.
func (c *CommunityVoteData) Decode(r io.Reader) error {
	numCandidates, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	c.ReplacementCandidates = make([][20]byte, numCandidates)
	for i := range c.ReplacementCandidates {
		if err := serializer.ReadElement(r, &c.ReplacementCandidates[i]); err != nil {
			return err
		}
	}

	return nil
}

// Copy copies the community vote data.
func (c *CommunityVoteData) Copy() *CommunityVoteData {
	newCommunityVoteData := *c
	newCommunityVoteData.ReplacementCandidates = make([][20]byte, len(c.ReplacementCandidates))
	copy(newCommunityVoteData.ReplacementCandidates, c.ReplacementCandidates)

	return &newCommunityVoteData
}

// Hash calculates the hash of the vote data.
func (c *CommunityVoteData) Hash() chainhash.Hash {
	buf := new(bytes.Buffer)
	c.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}

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
	VoteEpoch uint64
}

// Encode encodes the governance vote to a writer.
func (gv *GovernanceVote) Encode(w io.Writer) error {
	if err := serializer.WriteElements(w, gv.Type, gv.VoteEpoch); err != nil {
		return err
	}
	if err := serializer.WriteVarBytes(w, gv.Data); err != nil {
		return err
	}
	if err := bls.WriteFunctionalSignature(w, gv.Signature); err != nil {
		return err
	}
	return nil
}

// Decode decodes the governance vote from the reader.
func (gv *GovernanceVote) Decode(r io.Reader) error {
	if err := serializer.ReadElements(r, &gv.Type, &gv.VoteEpoch); err != nil {
		return err
	}
	data, err := serializer.ReadVarBytes(r)
	if err != nil {
		return err
	}
	gv.Data = data
	gv.Signature, err = bls.ReadFunctionalSignature(r)
	if err != nil {
		return err
	}

	return nil
}

func (gv *GovernanceVote) Valid() bool {
	sigHash := gv.SignatureHash()
	return gv.Signature.Verify(sigHash[:])
}

// SignatureHash gets the signed part of the hash.
func (gv *GovernanceVote) SignatureHash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	serializer.WriteVarBytes(buf, gv.Data)
	serializer.WriteElements(buf, gv.VoteEpoch, gv.Type)
	return chainhash.HashH(buf.Bytes())
}

// Hash calculates the hash of the governance vote.
func (gv *GovernanceVote) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = gv.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}

// Copy copies the governance vote.
func (gv *GovernanceVote) Copy() *GovernanceVote {
	newGv := *gv
	newGv.Data = make([]byte, len(gv.Data))
	copy(newGv.Data, gv.Data)
	newGv.Signature = gv.Signature.Copy()

	return &newGv
}
