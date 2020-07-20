package primitives

import (
	"bytes"

	fastssz "github.com/ferranbt/fastssz"
	"github.com/golang/snappy"
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

// GovernanceSerializable is a struct that contains the Governance state on a serializable struct.
type GovernanceSerializable struct {
	ReplaceVotes   []ReplacementVotes
	CommunityVotes []CommunityVoteDataInfo
}

// Governance is a struct that contains CommunityVotes and ReplacementVotes indexes and slices.
type Governance struct {
	ReplaceVotes   map[[20]byte]chainhash.Hash
	CommunityVotes map[chainhash.Hash]CommunityVoteData
	fastssz.Marshaler
	fastssz.Unmarshaler
}

// Marshal serializes the struct to bytes
func (g *Governance) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(g)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal deserialize the bytes to a struct
func (g *Governance) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(d, g)
}

// MarshalSSZ overrides the ssz function using fastssz interface
func (g *Governance) MarshalSSZ() ([]byte, error) {
	b := []byte{}
	return g.MarshalSSZTo(b)
}

// MarshalSSZTo utility function to override the ssz using the fastssz interface
func (g *Governance) MarshalSSZTo(dst []byte) ([]byte, error) {
	ser := g.ToSerializable()
	mb, err := ssz.Marshal(ser)
	if err != nil {
		return nil, err
	}
	copy(dst, mb)
	return mb, nil
}

// SizeSSZ utility function to override the ssz using the fastssz interface
func (g *Governance) SizeSSZ() int {
	total := 0
	total += len(g.CommunityVotes) * 132
	total += len(g.ReplaceVotes) * 52
	return total
}

// UnmarshalSSZ overrides the ssz function using the fastssz interface
func (g *Governance) UnmarshalSSZ(b []byte) error {
	ser := new(GovernanceSerializable)
	err := ssz.Unmarshal(b, ser)
	if err != nil {
		return err
	}
	g.FromSerializable(ser)
	return nil
}

// ToSerializable creates a copy of the struct into a slices struct
func (g *Governance) ToSerializable() GovernanceSerializable {
	replaceVotes := []ReplacementVotes{}
	communityVotes := []CommunityVoteDataInfo{}
	for k, v := range g.ReplaceVotes {
		replaceVotes = append(replaceVotes, ReplacementVotes{Account: k, Hash: v})
	}
	for k, v := range g.CommunityVotes {
		communityVotes = append(communityVotes, CommunityVoteDataInfo{Hash: k, Data: v})
	}
	return GovernanceSerializable{ReplaceVotes: replaceVotes, CommunityVotes: communityVotes}
}

// FromSerializable creates the struct into a map based struct
func (g *Governance) FromSerializable(s *GovernanceSerializable) {
	g.ReplaceVotes = map[[20]byte]chainhash.Hash{}
	g.CommunityVotes = map[chainhash.Hash]CommunityVoteData{}
	for _, v := range s.ReplaceVotes {
		g.ReplaceVotes[v.Account] = v.Hash
	}
	for _, v := range s.CommunityVotes {
		g.CommunityVotes[v.Hash] = v.Data
	}
	return
}

// CommunityVoteData is the votes that users sign to vote for a specific candidate.
type CommunityVoteData struct {
	ReplacementCandidates [][20]byte
}

// Marshal encodes the data.
func (c *CommunityVoteData) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(c)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
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
	b, err := ssz.Marshal(gv)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
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
