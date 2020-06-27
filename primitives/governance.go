package primitives

import (
	"bytes"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/csmt"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

type GovernanceState struct {
	// replaceVotes are votes to start the community-override functionality. Each address
	// in here must have at least 100 POLIS and once that accounts for >=30% of the supply,
	// a community voting round starts.
	// For a voting period, the hash is set to the proposed community vote.
	// For a non-voting period, the hash is 0.
	replaceVotes   *csmt.Tree
	communityVotes *csmt.Tree
	// CommunityVotes is set during a voting period to keep track of the
	// possible votes.
}

func (gs *GovernanceState) GetCommunityVote(hash chainhash.Hash) CommunityVoteData {
	return CommunityVoteData{}
}

func (gs *GovernanceState) SetCommunityVote(hash chainhash.Hash, vote CommunityVoteData) {
	return
}

func (gs *GovernanceState) SetReplaceVoteAccount(acc [20]byte, hash chainhash.Hash) {
	return
}

func (gs *GovernanceState) GetReplaceVoteAccount(acc [20]byte) (chainhash.Hash, bool) {
	return chainhash.Hash{}, false
}

func (gs *GovernanceState) GetReplaceVotes() map[[20]byte]chainhash.Hash {
	return map[[20]byte]chainhash.Hash{}
}

func (gs *GovernanceState) DeleteReplaceVoteAccount(acc [20]byte) {
	return
}

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
	hash, _ := ssz.HashTreeRoot(c)
	return chainhash.Hash(hash)
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
	Type          uint8
	Data          []byte
	FunctionalSig []byte
	VoteEpoch     uint64
}

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
	return ssz.Marshal(gv)
}

// Unmarshal decodes the data.
func (gv *GovernanceVote) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, gv)
}

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
	hash, _ := ssz.HashTreeRoot(gv)
	return chainhash.Hash(hash)
}

// Copy copies the governance vote.
func (gv *GovernanceVote) Copy() *GovernanceVote {
	newGv := *gv
	newGv.Data = make([]byte, len(gv.Data))
	copy(newGv.Data, gv.Data)
	newGv.FunctionalSig = gv.FunctionalSig
	return &newGv
}
