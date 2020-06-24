package primitives

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// GovernanceVoteType is the type of a governance vote.
type GovernanceVoteType uint8

// CommunityVoteData is the votes that users sign to vote for a specific candidate.
type CommunityVoteData struct {
	ReplacementCandidates [][20]byte
}

// Marshal encodes the data.
func (c *CommunityVoteData) Marshal() ([]byte, error) {
	return ssz.Marshal(c)
}

// Unmarshal decodes the data.
func (c *CommunityVoteData) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, c)
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
	b, _ := c.Marshal()
	return chainhash.HashH(b)
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

// Marshal encodes the data.
func (gv *GovernanceVote) Marshal() ([]byte, error) {
	return ssz.Marshal(gv)
}

// Unmarshal decodes the data.
func (gv *GovernanceVote) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, gv)
}

func (gv *GovernanceVote) Valid() bool {
	sigHash := gv.SignatureHash()
	return gv.Signature.Verify(sigHash[:])
}

// SignatureHash gets the signed part of the hash.
func (gv *GovernanceVote) SignatureHash() chainhash.Hash {
	cp := gv.Copy()
	cp.Signature = nil
	b, _ := cp.Marshal()
	return chainhash.HashH(b)
}

// Hash calculates the hash of the governance vote.
func (gv *GovernanceVote) Hash() chainhash.Hash {
	b, _ := gv.Marshal()
	return chainhash.HashH(b)
}

// Copy copies the governance vote.
func (gv *GovernanceVote) Copy() *GovernanceVote {
	newGv := *gv
	newGv.Data = make([]byte, len(gv.Data))
	copy(newGv.Data, gv.Data)
	newGv.Signature = gv.Signature.Copy()
	return &newGv
}
