package primitives

import (
	"bytes"
	"sync"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

type CommunityVoteDataInfo struct {
	Hash chainhash.Hash
	Data CommunityVoteData
}

type ReplacementVotes struct {
	Account [20]byte
	Hash    chainhash.Hash
}

var (
	repalceVotesLock      *sync.RWMutex
	communityVotesLock    *sync.RWMutex
	replacementVotesIndex map[[20]byte]int
	communityVotesIndex   map[chainhash.Hash]int
)

// Governance is a struct that contains CommunityVotes and ReplacementVotes indexes and slices.
type Governance struct {
	ReplacementVotes []ReplacementVotes
	CommunityVotes   []CommunityVoteDataInfo
}

// Load assumes the slices are filled and constuct the indexes.
func (g *Governance) Load() {
	for i, v := range g.ReplacementVotes {
		repalceVotesLock.Lock()
		replacementVotesIndex[v.Account] = i
		repalceVotesLock.Unlock()
	}
	for i, v := range g.CommunityVotes {
		communityVotesLock.Lock()
		communityVotesIndex[v.Hash] = i
		communityVotesLock.Unlock()
	}
}

func NewGovernanceState() Governance {
	return Governance{
		ReplacementVotes: []ReplacementVotes{},
		CommunityVotes:   []CommunityVoteDataInfo{},
	}
}

// GetAllReplacementVotes get all replacement votes on the state.
func (g *Governance) GetAllReplacementVotes() []ReplacementVotes {
	return g.ReplacementVotes
}

// GetReplacementVote returns a replacement vote.
func (g *Governance) GetReplacementVote(acc [20]byte) (chainhash.Hash, bool) {
	return chainhash.Hash{}, false
}

// SetReplacementVote sets a new replacement vote on the state
func (g *Governance) SetReplacementVote(acc [20]byte, hash chainhash.Hash) {
	return
}

// DeleteReplacementVote removes a vote on the slice and reconstruct index
func (g *Governance) DeleteReplacementVote(acc [20]byte) {
	return
}

// GetCommunityVote returns a vote on the community votes state.
func (g *Governance) GetCommunityVote(hash chainhash.Hash) CommunityVoteData {
	return CommunityVoteData{}
}

// SetCommunityVote stores a community vote on the state.
func (g *Governance) SetCommunityVote(hash chainhash.Hash, data CommunityVoteData) {
	return
}

func (g *Governance) getReplacementVoteIndex(acc [20]byte) (int, bool) {
	repalceVotesLock.Lock()
	defer repalceVotesLock.Unlock()
	i, ok := replacementVotesIndex[acc]
	return i, ok
}

func (g *Governance) getCommunityVoteIndex(hash chainhash.Hash) (int, bool) {
	communityVotesLock.Lock()
	defer communityVotesLock.Unlock()
	i, ok := communityVotesIndex[hash]
	return i, ok
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
