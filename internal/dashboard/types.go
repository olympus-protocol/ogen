package dashboard

type Data struct {
	NodeData          NodeData
	NetworkData       NetworkData
	KeystoreData      KeystoreData
	ProposerData      ProposerData
	PeerData          []PeerData
	ParticipationInfo ParticipationInfo
}

type NodeData struct {
	TipHeight       uint64
	TipSlot         uint64
	TipHash         string
	JustifiedHeight uint64
	JustifiedSlot   uint64
	JustifiedHash   string
	FinalizedHeight uint64
	FinalizedSlot   uint64
	FinalizedHash   string
}

type NetworkData struct {
	ID             string
	PeersConnected int
	PeersAhead     int
	PeersBehind    int
	PeersEqual     int
}

type KeystoreData struct {
	Keys              int
	Validators        int
	KeysParticipating int
}

type ProposerData struct {
	Slot      uint64
	Epoch     uint64
	Voting    bool
	Proposing bool
}

type PeerData struct {
	ID        string
	Finalized uint64
	Justified uint64
	Tip       uint64
}

type ParticipationInfo struct {
	EpochSlot               uint64
	Epoch                   uint64
	ParticipationPercentage string
}
