package primitives

import (
	"bytes"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// CommunityVoteDataInfo contains information about a community vote.
type CommunityVoteDataInfo struct {
	Hash [32]byte `ssz-size:"32"`
	Data *CommunityVoteData
}

// ReplacementVotes contains information about a replacement candidate selected.
type ReplacementVotes struct {
	Account [20]byte `ssz-size:"20"`
	Hash    [32]byte `ssz-size:"32"`
}

// GovernanceSerializable is a struct that contains the Governance state on a serializable struct.
type GovernanceSerializable struct {
	ReplaceVotes   []*ReplacementVotes      `ssz-max:"1099511627776"`
	CommunityVotes []*CommunityVoteDataInfo `ssz-max:"1099511627776"`
}

// CommunityVoteData is the votes that users sign to vote for a specific candidate.
type CommunityVoteData struct {
	ReplacementCandidates [5][20]byte `ssz-size:"20"`
}

// Marshal encodes the data.
func (c *CommunityVoteData) Marshal() ([]byte, error) {
	b, err := c.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (c *CommunityVoteData) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return c.UnmarshalSSZ(d)
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
	Data          []byte `ssz-size:"2048"` // TODO Calculate
	FunctionalSig []byte `ssz-size:"2048"` // TODO Calculate
	VoteEpoch     uint64
}

// Signature returns the governance vote bls signature.
func (gv *GovernanceVote) Signature() (bls.FunctionalSignature, error) {
	buf := bytes.NewBuffer([]byte{})
	buf.Write(gv.FunctionalSig)
	sig, err := bls.ReadFunctionalSignature(buf)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// Marshal encodes the data.
func (gv *GovernanceVote) Marshal() ([]byte, error) {
	b, err := gv.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (gv *GovernanceVote) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return gv.UnmarshalSSZ(d)
}

// Valid returns a boolean that checks for validity of the vote.
func (gv *GovernanceVote) Valid() bool {
	sigHash := gv.SignatureHash()
	sig, err := gv.Signature()
	if err != nil {
		return false
	}
	return sig.Verify(sigHash[:])
}

// SignatureHash gets the signed part of the hash.
func (gv *GovernanceVote) SignatureHash() chainhash.Hash {
	cp := gv.Copy()
	cp.FunctionalSig = []byte{}
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
	newGv.FunctionalSig = gv.FunctionalSig
	return &newGv
}
