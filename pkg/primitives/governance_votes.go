package primitives

import (
	"encoding/binary"
	"github.com/olympus-protocol/ogen/pkg/bls/multisig"

	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

const (
	// MaxReplacementsVoteSize is the maximum amount of bytes a ReplacementVotes can have
	MaxReplacementsVoteSize = 52
	// MaxGovernanceVoteSize is the maximum amount of bytes a GovernanceVote can have
	MaxGovernanceVoteSize = 116 + multisig.MaxMultisigSize
	// MaxCommunityVoteDataSize is the maximum amount of bytes a CommunityVoteData can have
	MaxCommunityVoteDataSize = 104
)

// CommunityVoteDataInfo contains information about a community vote.
type CommunityVoteDataInfo struct {
	Hash [32]byte `ssz-size:"32"`
	Data *CommunityVoteData
}

// ReplacementVotes contains information about a replacement candidate selected.
type ReplacementVotes struct {
	Account [20]byte
	Hash    [32]byte
}

// Marshal encodes the data.
func (r *ReplacementVotes) Marshal() ([]byte, error) {
	return r.MarshalSSZ()
}

// Unmarshal decodes the data.
func (r *ReplacementVotes) Unmarshal(b []byte) error {
	return r.UnmarshalSSZ(b)
}

// CommunityVoteData is the votes that users sign to vote for a specific candidate.
type CommunityVoteData struct {
	ReplacementCandidates [][20]byte `ssz-max:"5"`
}

// Marshal encodes the data.
func (c *CommunityVoteData) Marshal() ([]byte, error) {
	return c.MarshalSSZ()
}

// Unmarshal decodes the data.
func (c *CommunityVoteData) Unmarshal(b []byte) error {
	return c.UnmarshalSSZ(b)
}

// Copy copies the community vote data.
func (c *CommunityVoteData) Copy() *CommunityVoteData {
	newCommunityVoteData := &CommunityVoteData{
		ReplacementCandidates: make([][20]byte, len(c.ReplacementCandidates)),
	}
	for i := range c.ReplacementCandidates {
		newCommunityVoteData.ReplacementCandidates[i] = c.ReplacementCandidates[i]
	}
	return newCommunityVoteData
}

// Hash calculates the hash of the vote data.
func (c *CommunityVoteData) Hash() chainhash.Hash {
	b, _ := c.Marshal()
	return chainhash.HashH(b)
}

const (
	// EnterVotingPeriod can be done by anyone on the network to signal that they
	// want a voting period to start.
	EnterVotingPeriod uint64 = iota

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
	Type      uint64
	Data      [100]byte
	Multisig  *multisig.Multisig
	VoteEpoch uint64
}

// Marshal encodes the data.
func (g *GovernanceVote) Marshal() ([]byte, error) {
	return g.MarshalSSZ()
}

// Unmarshal decodes the data.
func (g *GovernanceVote) Unmarshal(b []byte) error {
	return g.UnmarshalSSZ(b)
}

// Valid returns a boolean that checks for validity of the vote
func (g *GovernanceVote) Valid() bool {
	sigHash := g.SignatureHash()
	return g.Multisig.Verify(sigHash[:])
}

// SignatureHash gets the signed part of the hash.
func (g *GovernanceVote) SignatureHash() chainhash.Hash {
	buf := make([]byte, 116)
	copy(buf[:], g.Data[:])
	binary.LittleEndian.PutUint64(buf, g.Type)
	binary.LittleEndian.PutUint64(buf, g.VoteEpoch)
	return chainhash.HashH(buf)
}

// Hash calculates the hash of the governance vote.
func (g *GovernanceVote) Hash() chainhash.Hash {
	b, _ := g.Marshal()
	return chainhash.HashH(b)
}

// Copy copies the governance vote.
func (g *GovernanceVote) Copy() *GovernanceVote {
	newGv := &GovernanceVote{
		Type:      g.Type,
		VoteEpoch: g.VoteEpoch,
		Data:      [100]byte{},
	}
	copy(newGv.Data[:], g.Data[:])
	if g.Multisig != nil {
		newGv.Multisig = g.Multisig.Copy()
	}
	return newGv
}
