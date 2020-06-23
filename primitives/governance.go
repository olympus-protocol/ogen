package primitives

import (
	"encoding/binary"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type CommunityVoteData struct {
	ReplacementCandidates [][]byte `ssz-size:"?,20" ssz-max:"1099511627776"`
}

// Copy copies the community vote data.
func (c *CommunityVoteData) Copy() *CommunityVoteData {
	newCommunityVoteData := *c
	newCommunityVoteData.ReplacementCandidates = make([][]byte, len(c.ReplacementCandidates))
	copy(newCommunityVoteData.ReplacementCandidates, c.ReplacementCandidates)
	return &newCommunityVoteData
}

// Hash calculates the hash of the vote data.
func (c *CommunityVoteData) Hash() chainhash.Hash {
	// TODO handle error
	b, _ := c.MarshalSSZ()
	return chainhash.HashH(b)
}

const (
	// EnterVotingPeriod can be done by anyone on the network to signal that they
	// want a voting period to start.
	EnterVotingPeriod uint8 = iota

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
	Type      uint8
	Data      []byte `ssz-size:"20"`
	Signature []byte `ssz-size:"96"`
	VoteEpoch uint64
}

func (gv *GovernanceVote) Valid() bool {
	sigHash := gv.SignatureHash()
	sig, err := bls.SignatureFromBytes(gv.Signature)
	if err != nil {
		return false
	}
	return sig.Verify(sigHash[:])
}

func (gv *GovernanceVote) GetSignature() *bls.FunctionalSignature {

}

// SignatureHash gets the signed part of the hash.
func (gv *GovernanceVote) SignatureHash() chainhash.Hash {
	buf := []byte{}
	copy(buf, gv.Data)
	binary.LittleEndian.PutUint64(buf, gv.VoteEpoch)
	buf = append(buf, gv.Type)
	return chainhash.HashH(buf)
}

// Hash calculates the hash of the governance vote.
func (gv *GovernanceVote) Hash() chainhash.Hash {
	b, _ := gv.MarshalSSZ()
	return chainhash.HashH(b)
}

// Copy copies the governance vote.
func (gv *GovernanceVote) Copy() *GovernanceVote {
	newGv := *gv
	newGv.Data = make([]byte, len(gv.Data))
	copy(newGv.Data, gv.Data)
	copy(newGv.Signature, gv.Signature)
	return &newGv
}
