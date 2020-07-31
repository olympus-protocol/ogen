package primitives

import (
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// Governance is a struct that contains CommunityVotes and ReplacementVotes indexes and slices.
type Governance struct {
	ReplaceVotes   map[[20]byte]chainhash.Hash
	CommunityVotes map[chainhash.Hash]CommunityVoteData
}

// Marshal serializes the struct to bytes
func (g *Governance) Marshal() ([]byte, error) {
	s := g.ToSerializable()
	b, err := s.MarshalSSZ()
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
	gs := new(GovernanceSerializable)
	err = gs.UnmarshalSSZ(d)
	if err != nil {
		return err
	}
	g.FromSerializable(gs)
	return nil
}

// ToSerializable creates a copy of the struct into a slices struct
func (g *Governance) ToSerializable() GovernanceSerializable {
	replaceVotes := []*ReplacementVotes{}
	communityVotes := []*CommunityVoteDataInfo{}
	for k, v := range g.ReplaceVotes {
		replaceVotes = append(replaceVotes, &ReplacementVotes{Account: k, Hash: v})
	}
	for k, v := range g.CommunityVotes {
		communityVotes = append(communityVotes, &CommunityVoteDataInfo{Hash: k, Data: &v})
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
		g.CommunityVotes[v.Hash] = *v.Data
	}
	return
}
