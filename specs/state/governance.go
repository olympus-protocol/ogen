package state

type CommunityVoteData struct {
	ReplacementCandidates [][]byte `ssz-size:"?,20" ssz-max:"1099511627776"`
}

type GovernanceVote struct {
	Type      uint8
	Data      []byte `ssz-size:"20"`
	Signature []byte `ssz-size:"96"`
	VoteEpoch uint64
}
