package primitives

import (
	"bytes"
	"errors"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var (
	// ErrorReplacementsVoteSize is returned when a ReplacementVotes exceed MaxReplacementVoteSize
	ErrorReplacementsVoteSize = errors.New("replacement vote size too big")
	// ErrorGovernanceVote is returned when a GovernanceVote exceed MaxGovernanceVoteSize
	ErrorGovernanceVoteSize = errors.New("governance vote size too big")
	// ErrorCommunityVoteData is returned when a CommunityVoteData exceed MaxCommunityVoteDataSize
	ErrorCommunityVoteDataSize = errors.New("community vote size too big")
)

const (
	// MaxReplacementsVoteSize is the maximum amount of bytes a ReplacementVotes can have
	MaxReplacementsVoteSize = 52
	// MaxGovernanceVoteSize is the maximum amount of bytes a GovernanceVote can have
	MaxGovernanceVoteSize = 260
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
	b, err := r.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxReplacementsVoteSize {
		return nil, ErrorReplacementsVoteSize
	}
	return b, nil
}

// Unmarshal decodes the data.
func (r *ReplacementVotes) Unmarshal(b []byte) error {
	if len(b) > MaxReplacementsVoteSize {
		return ErrorReplacementsVoteSize
	}
	return r.UnmarshalSSZ(b)
}

// GovernanceSerializable is a struct that contains the Governance state on a serializable struct.
type GovernanceSerializable struct {
	ReplaceVotes   []*ReplacementVotes      `ssz-max:"1099511627776"`
	CommunityVotes []*CommunityVoteDataInfo `ssz-max:"1099511627776"`
}

// CommunityVoteData is the votes that users sign to vote for a specific candidate.
type CommunityVoteData struct {
	ReplacementCandidates [][20]byte `ssz-max:"5"`
}

// Marshal encodes the data.
func (c *CommunityVoteData) Marshal() ([]byte, error) {
	b, err := c.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxCommunityVoteDataSize {
		return nil, ErrorCommunityVoteDataSize
	}
	return b, nil
}

// Unmarshal decodes the data.
func (c *CommunityVoteData) Unmarshal(b []byte) error {
	if len(b) > MaxCommunityVoteDataSize {
		return ErrorCommunityVoteDataSize
	}
	return c.UnmarshalSSZ(b)
}

// Copy copies the community vote data.
func (c *CommunityVoteData) Copy() *CommunityVoteData {
	newCommunityVoteData := *c
	newCommunityVoteData.ReplacementCandidates = c.ReplacementCandidates
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
	Type          uint64
	Data          [100]byte
	FunctionalSig [144]byte
	VoteEpoch     uint64
}

// Signature returns the governance vote bls signature.
func (g *GovernanceVote) Signature() (bls.FunctionalSignature, error) {
	buf := bytes.NewBuffer([]byte{})
	buf.Write(g.FunctionalSig[:])
	sig, err := bls.ReadFunctionalSignature(buf)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// Marshal encodes the data.
func (g *GovernanceVote) Marshal() ([]byte, error) {
	b, err := g.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxGovernanceVoteSize {
		return nil, ErrorGovernanceVoteSize
	}
	return b, nil
}

// Unmarshal decodes the data.
func (g *GovernanceVote) Unmarshal(b []byte) error {
	if len(b) > MaxGovernanceVoteSize {
		return ErrorGovernanceVoteSize
	}
	return g.UnmarshalSSZ(b)
}

// Valid returns a boolean that checks for validity of the vote.
func (g *GovernanceVote) Valid() bool {
	sigHash := g.SignatureHash()
	sig, err := g.Signature()
	if err != nil {
		return false
	}
	return sig.Verify(sigHash[:])
}

// SignatureHash gets the signed part of the hash.
func (g *GovernanceVote) SignatureHash() chainhash.Hash {
	cp := g.Copy()
	cp.FunctionalSig = [144]byte{}
	b, _ := cp.Marshal()
	return chainhash.HashH(b)
}

// Hash calculates the hash of the governance vote.
func (g *GovernanceVote) Hash() chainhash.Hash {
	b, _ := g.Marshal()
	return chainhash.HashH(b)
}

// Copy copies the governance vote.
func (g *GovernanceVote) Copy() *GovernanceVote {
	newGv := *g
	newGv.Data = [100]byte{}
	copy(newGv.Data[:], g.Data[:])
	newGv.FunctionalSig = g.FunctionalSig
	return &newGv
}
