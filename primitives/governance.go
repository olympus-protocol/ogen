package primitives

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// GovernanceVoteType is the type of a governance vote.
type GovernanceVoteType uint8

// CommunityVoteData is the votes that users sign to vote for a specific candidate.
type CommunityVoteData struct {
	ReplacementCandidates [][]byte
}

// Copy copies the community vote data.
func (c *CommunityVoteData) Copy() *CommunityVoteData {
	newCommunityVoteData := *c
	newCommunityVoteData.ReplacementCandidates = make([][]byte, len(c.ReplacementCandidates))
	copy(newCommunityVoteData.ReplacementCandidates, c.ReplacementCandidates)

	return &newCommunityVoteData
}

// Hash calculates the hash of the vote data.
func (c *CommunityVoteData) Hash() (chainhash.Hash, error) {
	ser, err := c.Marshal()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(ser), nil
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
	Signature []byte
	VoteEpoch uint64
}

func (gv *GovernanceVote) Valid() bool {
	sigHash := gv.SignatureHash()
	return gv.Signature.Verify(sigHash[:])
}

// SignatureHash gets the signed part of the hash.
func (gv *GovernanceVote) SignatureHash() chainhash.Hash {
	buf := []byte{}
	// TODO
	//serializer.WriteVarBytes(buf, gv.Data)
	//serializer.WriteElements(buf, gv.VoteEpoch, gv.Type)
	return chainhash.HashH(buf)
}

// Hash calculates the hash of the governance vote.
func (gv *GovernanceVote) Hash() (chainhash.Hash, error) {
	ser, err := gv.Marshal()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(ser), nil
}

// Copy copies the governance vote.
func (gv *GovernanceVote) Copy() *GovernanceVote {
	newGv := *gv
	newGv.Data = make([]byte, len(gv.Data))
	copy(newGv.Data, gv.Data)
	newGv.Signature = gv.Signature.Copy()

	return &newGv
}
